name: "database-query"
description: "Execute database queries safely"
parameters:
  type: object
  properties:
    query:
      type: string
      description: "The SQL query to execute"
    params:
      type: array
      description: "Query parameters"
  required: ["query"]

execute:
  command: "node"
  args: ["scripts/db-query.js", "{{query}}"]
  env:
    DATABASE_URL: "{{.DatabaseURL}}"

permissions:
  - "database:read"
  - "database:write"