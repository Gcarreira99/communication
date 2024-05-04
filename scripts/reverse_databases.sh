
rm -rf $HOME/backup/data/databases/neo4j/*
rm -rf $HOME/backup/data/transactions/neo4j/*

cp -r --verbose ../../../neo4j/data/databases/twitter-v2-50/*  ../../../backup/data/databases/neo4j
cp -r --verbose ../../../neo4j/data/transactions/twitter-v2-50/* ../../../backup/data/transactions/neo4j

rm -rf $HOME/neo4j/data/databases/neo4j/*
rm -rf $HOME/neo4j/data/transactions/neo4j/*

cp -r --verbose ../../../neo4j/data/databases/twitter-v2-50/* ../../../neo4j/data/databases/neo4j
cp -r --verbose ../../../neo4j/data/transactions/twitter-v2-50/* ../../../neo4j/data/transactions/neo4j
