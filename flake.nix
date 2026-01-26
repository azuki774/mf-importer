{
  description = "mf-importer development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Pin Go to specific version 1.25.6
        go = pkgs.go_1_25.overrideAttrs (oldAttrs: rec {
          version = "1.25.6";
          src = pkgs.fetchurl {
            url = "https://go.dev/dl/go${version}.src.tar.gz";
            hash = "sha256-WMv3ceRNdt5vVtGeM7d9dFoeSJNAkih15GWFuXXCsFk=";
          };
        });
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go development
            go  # Use pinned version
            gopls
            gotools
            go-tools # staticcheck
            golangci-lint
            
            # Node.js development
            nodejs_22
            nodePackages.npm
            
            # Database tools
            mariadb.client
            
            # Docker
            docker
            docker-compose
            
            # Version control and CI/CD
            git
            gh
            
            # Shell utilities
            zsh
            
            # Build tools
            gnumake
            
            # SQL migration tool
            # Note: sql-migrate is in vendor_ci/, so we don't need it from nixpkgs
          ];

          shellHook = ''
            echo "🚀 mf-importer development environment loaded"
            echo "📦 Go version: $(go version)"
            echo "📦 Node version: $(node --version)"
            echo "📦 npm version: $(npm --version)"
            echo ""
            echo "💡 Run 'make migration' to run database migrations"
            echo "💡 Run 'make start' to start all services with docker-compose"
            echo "💡 Run 'cd frontend && npm run dev' to start frontend development server"
            
            # Set up Go environment
            export GOPATH=$HOME/go
            export PATH=$GOPATH/bin:$PATH
            
            # Set up database connection defaults
            export DB_HOST=localhost
            export DB_PORT=3306
            export DB_USER=root
            export DB_PASS=password
            export DB_NAME=mfimporter
          '';
        };
      }
    );
}
