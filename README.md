# Dev Setup

1. Download https://iptoasn.com/data/ip2asn-combined.tsv.gz and place it in the `src` directory
2. Go into `src` directory and run `go run .`

# Deployment

1. `./production-build.sh`
2. `mv out ip-atlas`
3. `scp -r ip-atlas/ SERVER_USER@SERVER_IP:/www`
4. `mv ip-atlas out`
5. Make sure you have cron setup `0 0 1 * * /usr/bin/bash -c "cd /www/ip-atlas/ && IP_ATLAS_PRODUCTION=TRUE ./ip-atlas.exe"`
