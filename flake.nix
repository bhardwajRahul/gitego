{
  description = "git-ego - Git identity manager with directory-based profile switching";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
          git-ego = pkgs.buildGoModule {
            pname = "git-ego";
            version = "0.2.4";
            src = self;
            vendorHash = "sha256-95hogC0JlrnUJBvoV4Ac4F5XmOqsQh2AQ0rZGcC/4oY=";
            ldflags = [ "-s" "-w" "-X github.com/bgreenwell/git-ego/cmd.version=0.2.4" ];
            nativeBuildInputs = [ pkgs.git ];

            meta = with pkgs.lib; {
              description = "Git identity manager with directory-based profile switching";
              homepage = "https://github.com/bgreenwell/git-ego";
              license = licenses.mit;
              mainProgram = "git-ego";
              platforms = platforms.unix;
            };
          };
        in
        {
          default = git-ego;
          inherit git-ego;
        });

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.git-ego}/bin/git-ego";
        };
      });
    };
}
