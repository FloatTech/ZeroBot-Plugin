{
  pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
      import (fetchTree nixpkgs.locked) {
        overlays = [
          (import "${fetchTree gomod2nix.locked}/overlay.nix")
        ];
      }
  ),
  buildGoApplication ? pkgs.buildGoApplication,
  ...
}:
buildGoApplication {
  pname = "ZeroBot-Plugin";
  version = "1.8.0";
  pwd = ./.;
  src = ./.;
  go = pkgs.go_1_24;
  preBuild = ''
    go generate main.go
  '';
  modules = ./gomod2nix.toml;
}
