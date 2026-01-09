name: "deploy"
description: "Deploy the application to production"
aliases: ["prod-deploy"]

steps:
  - name: "Run tests"
    command: "npm test"
  - name: "Build application"
    command: "npm run build"
  - name: "Deploy to server"
    command: "rsync -avz dist/ user@server:/var/www/app/"
  - name: "Restart services"
    command: "ssh user@server 'sudo systemctl restart app'"

env:
  NODE_ENV: "production"