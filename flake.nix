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
