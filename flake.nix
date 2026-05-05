{
  description = "Zero-config Cloudflare Tunnel automation for Docker";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable-small";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "dev";
        commit = if (self ? rev) then self.rev else "dev";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "zero-tunnel";
          inherit version;
          src = ./.;

          vendorHash = "sha256-Z6b1yoRzB3UEwCjDPiMaQ/2z5MNMPRQ07C9ykajQ1Fg=";

          env.CGO_ENABLED = 0;

          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${version}"
            "-X main.Commit=${commit}"
            "-X main.BuildDate=nix-build"
          ];

          doCheck = false; # Skip tests during build as they need docker
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            air
            go
            gopls
            goreleaser
          ];

          shellHooks = ''
            echo "Zero-tunnel development environment"
          '';
        };
      }
    );
}
