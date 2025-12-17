# Step 1: Build React

FROM node:20-alpine3.20 AS build

WORKDIR /app

RUN apk update && apk upgrade --no-cache

RUN corepack enable

COPY package.json yarn.lock ./

RUN yarn install --frozen-lockfile

COPY . .

RUN yarn build

# Step 2: Serve with Nginx
FROM nginx:alpine-slim AS production


RUN apk update && apk upgrade --no-cache

COPY --from=build /app/dist /usr/share/nginx/html

COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]