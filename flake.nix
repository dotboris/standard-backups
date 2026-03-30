{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      formatter = pkgs.writeShellScriptBin "fmt" ''
        ${pkgs.alejandra}/bin/alejandra .
      '';

      devShells.default = pkgs.mkShell {
        # This old version of restic doesn't return the right status code when a
        # repository is not found. We want to support it too because it's still
        # in use in ubuntu 24.04.
        RESTIC_0_16 = pkgs.restic.overrideAttrs (final: previous: rec {
          version = "0.16.5";
          src = pkgs.fetchFromGitHub {
            owner = "restic";
            repo = "restic";
            rev = "v${version}";
            hash = "sha256-WwySXQU8eoyQRcI+zF+pIIKLEFheTnqkPTw0IZeUrhA=";
          };
          vendorHash = "sha256-VZTX0LPZkqN4+OaaIkwepbGwPtud8Cu7Uq7t1bAUC8M=";
        });
        hardeningDisable = [
          "fortify" # required to use delve debugger
        ];
        packages = [
          # Nix & flake
          pkgs.nil
          pkgs.alejandra

          # Go
          pkgs.go
          pkgs.cobra-cli
          pkgs.go-mockery

          # Misc tooling
          pkgs.golangci-lint
          pkgs.goreleaser
          pkgs.svu
          pkgs.gotestsum
          pkgs.just
          pkgs.rsync
          pkgs.restic
          pkgs.bash # used in e2e tests
        ];
      };
    });
}
