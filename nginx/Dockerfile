# Use the official Nginx image as the base image
FROM nginx:alpine

# Copy the Nginx configuration file
COPY nginx/default.conf /etc/nginx/conf.d/default.conf

# Remove default nginx index page
RUN rm -rf /usr/share/nginx/html/*

# Copy static files from frontend build
COPY frontend/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Start Nginx when the container runs
CMD ["nginx", "-g", "daemon off;"]