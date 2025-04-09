import axios from "axios";
import querystring from "querystring";
import SpotifyWebApi from "spotify-web-api-node";

// Define Spotify client info and base API
const spotifyApi = new SpotifyWebApi({
  clientId: process.env.SPOTIFY_CLIENT_ID,
  clientSecret: process.env.SPOTIFY_CLIENT_SECRET,
});

const spotifyAPI = "https://api.spotify.com/v1/tracks/";

// Helper function to refresh the access token
async function refreshAccessToken(refresh_token) {
  const response = await axios.post(
    "https://accounts.spotify.com/api/token",
    querystring.stringify({
      grant_type: "refresh_token",
      refresh_token: refresh_token,
    }),
    {
      headers: {
        Authorization:
          "Basic " +
          Buffer.from(
            `${process.env.SPOTIFY_CLIENT_ID}:${process.env.SPOTIFY_CLIENT_SECRET}`
          ).toString("base64"),
        "Content-Type": "application/x-www-form-urlencoded",
      },
    }
  );

  return response.data.access_token;
}

export default async function handler(req, res) {
  // Extract the access token and refresh token from cookies
  const { access_token, refresh_token } = req.cookies;

  // If no access token, return unauthorized error
  if (!access_token) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  // Check if the access token is expired and refresh if necessary
  try {
    spotifyApi.setAccessToken(access_token);

    // Try making a Spotify API call to see if the token is valid
    await spotifyApi.getMe(); // This is a simple API call to check token validity
  } catch (error) {
    if (error.statusCode === 401 && refresh_token) {
      // If the token is expired, refresh it
      const newAccessToken = await refreshAccessToken(refresh_token);
      // Set the new access token
      spotifyApi.setAccessToken(newAccessToken);

      // Optionally, update the cookies with the new token
      res.setHeader(
        "Set-Cookie",
        `access_token=${newAccessToken}; HttpOnly; Path=/`
      );
    } else {
      return res.status(500).json({ error: "Authentication failed" });
    }
  }

  // Now that the token is valid, proceed with the API call for track data
  try {
    const trackId = req.query.trackId; // Assuming you're sending trackId as query parameter
    if (!trackId) {
      return res.status(400).json({ error: "Track ID is required" });
    }

    // Fetch track details by ID, including popularity
    const { body } = await spotifyApi.getTrack(trackId);

    // Extract track popularity and other details
    const trackDetails = {
      name: body.name,
      popularity: body.popularity,
      artist: body.artists[0].name,
      album: body.album.name,
      preview_url: body.preview_url, // Optional, track preview URL
      uri: body.uri, // Spotify URI for the track
    };

    // Send the track details (including popularity) back to the frontend
    res.status(200).json(trackDetails);
  } catch (error) {
    console.error("Error fetching track data:", error);
    res.status(500).json({ error: "Failed to fetch track data" });
  }
}
