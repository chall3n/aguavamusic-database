import SpotifyWebApi from "spotify-web-api-node";

const spotifyApi = new SpotifyWebApi({
  clientId: process.env.SPOTIFY_CLIENT_ID,
  clientSecret: process.env.SPOTIFY_CLIENT_SECRET,
});

export default async function handler(req, res) {
  // Extract the access token from cookies
  const { access_token } = req.cookies;

  // If no access token, return unauthorized error
  if (!access_token) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  console.log("Access token", access_token);
  // Set the access token to the Spotify API client
  spotifyApi.setAccessToken(access_token);

  try {
    // You can either fetch a specific track by its ID or fetch currently playing track
    // For example, to get details for a specific track by ID:
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
