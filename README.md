Music Database App

Full-stack web app using Next.js and Go made to display statistics for my personal music catalog with 
sorting functionality. Integrated with the Spotify API for dynamically updating "track popularity", 
an important metric Spotify uses internally. 

Frontend - Next.js, React.js, Tailwind CSS
Backend - Go, Gin
API - Spotify Web API
Database - Currently using Go arrays to simulate a database 

Make sure you have the following installed
-Node.js
-Go
-npm 

Setup & Installation
1. Clone the repository
2. Backend setup ('API' folder) navigate to the API directory, create .env file, edit to add your actual
   Spotify "CLIENT_ID" and "CLIENT_SECRET" (see instructions below)
3. Frontend setup ('aguava' folder) navigate to the frontend directory, and install Node.js dependencies

   You need to run both the backend and frontend servers simultaneously

Setting up Spotify API Access

This app uses the Spotify API to fetch real-time track popularity data. 
To run it locally, you'll need your own Spotify API credentials

1.  Go to (https://developer.spotify.com/dashboard/) and log in or create an account.
2.  Create a new application to get your Client ID and Client Secret.
3.  Copy the `.env.example` file in the `API/` directory to a new file named `.env`:
    ```bash
    cp API/.env.example API/.env
    ```
4.  Paste your Client ID and Client Secret into the `API/.env` file.
