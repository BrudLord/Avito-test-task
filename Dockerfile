FROM ubuntu:latest
LABEL authors="brudlord"

ENTRYPOINT ["top", "-b"]