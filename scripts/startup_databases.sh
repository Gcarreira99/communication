#Primary database
docker start testneo4j

#Verifier database
docker run --rm -d --publish=7475:7474 --publish=7688:7687 --volume=$HOME/backup/data:/data --volume=$HOME/backup/logs:/logs neo4j:latest
