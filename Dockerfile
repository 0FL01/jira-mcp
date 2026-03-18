FROM gcr.io/distroless/static-debian12:nonroot
COPY jira-mcp /jira-mcp
ENTRYPOINT ["/jira-mcp"]
