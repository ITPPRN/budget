# Step 1: Build React
FROM node:22-alpine AS build

WORKDIR /app

RUN apk update && apk upgrade --no-cache

RUN corepack enable

# Yarn Berry ต้องการ .yarnrc.yml และ .yarn/releases
COPY package.json yarn.lock .yarnrc.yml ./
COPY .yarn .yarn

RUN yarn install --immutable

COPY . .

RUN yarn build

# Step 2: Serve with Nginx
FROM nginx:alpine-slim AS production

RUN apk update && apk upgrade --no-cache

COPY --from=build /app/dist /usr/share/nginx/html

# COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY nginx/nginx.conf /etc/nginx/templates/default.conf.template

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
