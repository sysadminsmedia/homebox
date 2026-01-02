{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=master";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default =
          with pkgs;
          let
            go = go_1_24;
          in
          pkgs.mkShell {
            buildInputs = [
              # generic
              go-task

              # frontend
              nodejs_22
              pnpm
              nodePackages.typescript
              nodePackages.typescript-language-server

              # backend
              go
              glibc.static
              gofumpt
              golangci-lint
              sqlite
              go-swag
              gcc
              libwebp
              libavif
              libheif
              libjxl
            ];
            CFLAGS = "-I${pkgs.glibc.dev}/include";
            LDFLAGS = "-L${pkgs.glibc}/lib";
            GO = "${go}/bin/go";
            GOROOT = "${go}/share/go";
          };
      }
    );
}
