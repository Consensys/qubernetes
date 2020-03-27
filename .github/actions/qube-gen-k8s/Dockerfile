FROM quorumengineering/qubernetes

# we don't want to use the generators baked into the docker image,
# as it is the generators that we need to test: copy the committed base directory
# over the to the base docker image directory.
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
