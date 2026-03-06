FROM ubuntu:latest
LABEL authors="aleksejmajka"

ENTRYPOINT ["top", "-b"]