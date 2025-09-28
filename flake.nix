{
  description = "devx - A tool for generating, validating & sharing all your configurations";

  inputs.nixpkgs.url = "nixpkgs/nixos-25.05";

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
          version = "v0.4.39";
        in
        {
          default = pkgs.buildGoModule {
            inherit version;
            pname = "devx";
            src = pkgs.fetchFromGitHub {
              owner = "stakpak";
              repo = "devx";
              rev = version;
              sha256 = "sha256-uvDWGI9tj7Th9NFNzy2Vtb+Ekr+dz8bIIdasodf0oQA=";
            };
            subPackages = [
              "cmd/devx"
            ];
            vendorHash = "sha256-nHSV6yH2lYUg1wKgq2i/CBz0MVYbldnZ/pN0s6QQ2t4=";

            meta = with pkgs.lib; {
              description = ''
                A tool for generating, validating & sharing all your configurations,
                powered by CUE. Works with Kubernetes, Terraform, Compose, GitHub actions and much more...
              '';
              homepage = "https://devx.stakpak.dev/";
              license = licenses.asl20;
              platforms = platforms.linux;
            };
            tags = [
              "devx"
              "cue"
              "Kubernetes"
            ];
          };
          docker = pkgs.dockerTools.buildImage {
            name = "devx";
            tag = version;
            copyToRoot = pkgs.buildEnv {
              name = "image-root";
              paths = [ self.packages.${system}.default ];
              pathsToLink = [ "/bin" ];
            };
            config = {
              Cmd = [ "/bin/devx" ];
            };
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              gotools
              go-tools
            ];
          };
        }
      );
    };
}
