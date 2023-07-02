# Why?

While browsing the internet one day, I came across [iptoasn.com](https://www.iptoasn.com). This prompted two questions in my mind: "Who owns how many IPs?" and "How many IPs are not yet routed?". Curious to find out, I decided to investigate. The results surprised me because I had assumed that a larger percentage of IPv6 addresses would still be un-routed. However, I discovered that both IPv4 and IPv6 addresses had a similar percentage of un-routed IPs. Since this was more of a "research" projet, I didn't really care about the UI (and the result shows), so please, don't criticize me for that :D. Feel free to check it out and/or use the provided API.

# Dev Setup

1. Download https://iptoasn.com/data/ip2asn-combined.tsv.gz and place it in the `src` directory
2. Go into `src` directory and run `go run .`

# Deployment

1. `./production-build.sh`
2. `mv out ip-atlas`
3. `scp -r ip-atlas/ SERVER_USER@SERVER_IP:/www`
4. `mv ip-atlas out`
5. Make sure you have cron setup `0 0 1 * * /usr/bin/bash -c "cd /www/ip-atlas/ && IP_ATLAS_PRODUCTION=TRUE ./ip-atlas.exe"`
