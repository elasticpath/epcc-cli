FROM rodolpheche/wiremock:2.31.0
RUN apt-get update && apt-get install -y reflex
ADD https://repo1.maven.org/maven2/com/opentable/wiremock-body-transformer/1.1.6/wiremock-body-transformer-1.1.6.jar /var/wiremock/extensions/

ADD reflex-entrypoint.sh /
ENTRYPOINT /reflex-entrypoint.sh
