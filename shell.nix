{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
    buildInputs = [
        pkgs.go
        pkgs.gcc
    ];

    shellHook = ''
        export GOPATH=$(pwd)/go
        export PATH=$GOPATH/bin:$PATH
    '';
}