{
  description = "基于 ZeroBot 的 OneBot 插件";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";
  inputs.gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
  inputs.gomod2nix.inputs.flake-utils.follows = "flake-utils";

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    gomod2nix,
  }: (flake-utils.lib.eachDefaultSystem (
    system: let
      pkgs = nixpkgs.legacyPackages.${system};

      callPackage = pkgs.callPackage;
      # Simple test check added to nix flake check
      go-test = pkgs.stdenvNoCC.mkDerivation {
        name = "go-test";
        dontBuild = true;
        src = ./.;
        doCheck = true;
        nativeBuildInputs = with pkgs; [
          go
          writableTmpDirAsHomeHook
        ];
        checkPhase = ''
          go test -v ./...
        '';
        installPhase = ''
          mkdir "$out"
        '';
      };
      # Simple lint check added to nix flake check
      go-lint = pkgs.stdenvNoCC.mkDerivation {
        name = "go-lint";
        dontBuild = true;
        src = ./.;
        doCheck = true;
        nativeBuildInputs = with pkgs; [
          golangci-lint
          go
          writableTmpDirAsHomeHook
        ];
        checkPhase = ''
          golangci-lint run
        '';
        installPhase = ''
          mkdir "$out"
        '';
      };
      # doCheck will fail at download files
      ZeroBot-Plugin = (callPackage ./. {
        inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
      }).overrideAttrs (_: {doCheck = false;});
      # Build container layered image, useful overtime to save storage on duplicated layers
      containerImage = pkgs.dockerTools.buildLayeredImage {
        name = "ZeroBot-Plugin";
        tag = "latest";
        created = "now";
        contents = [
          pkgs.cacert
          pkgs.openssl
        ];
        config = {
          Cmd = ["${ZeroBot-Plugin}/bin/ZeroBot-Plugin"];
        };
      };
    in {
      inherit containerImage;
      checks = {
        inherit go-test go-lint;
      };
      packages.default = ZeroBot-Plugin;
      devShells.default = callPackage ./shell.nix {
        inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
      };
      # Custom application to build and load container image into the docker daemon
      # For now docker is a requirement
      apps.build-and-load = {
        type = "app";
        program = "${pkgs.writeShellScriptBin "build-and-load" ''
          nix build .#containerImage.${system}
          docker load < result
          echo "Container image loaded"
        ''}/bin/build-and-load";
      };
      formatter = pkgs.alejandra;
    }
  ));
}
