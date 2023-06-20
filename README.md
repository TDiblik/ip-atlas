# Dev Setup

1. Download https://iptoasn.com/data/ip2asn-combined.tsv.gz and place it in the `src` directory
2. Go into `src` directory and run `go run .`

# Deployment

1. `./production-build.sh`
2. `scp -r tomasdiblik.cz/ SERVER_USER@SERVER_IP:/www`
