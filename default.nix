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
}:
buildGoApplication {
  pname = "ZeroBot-Plugin";
  version = "1.7.6";
  pwd = ./.;
  src = ./.;
  # spec go version manually bcs
  # https://github.com/nix-community/gomod2nix/blob/30e3c3a9ec4ac8453282ca7f67fca9e1da12c3e6/builder/default.nix#L130
  # do no work
  go = pkgs.go_1_20;
  modules = ./gomod2nix.toml;
}
