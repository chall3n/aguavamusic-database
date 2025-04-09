import { NextResponse } from "next/server";
import querystring from "querystring";

const client_id = process.env.SPOTIFY_CLIENT_ID;
const redirect_uri = process.env.SPOTIFY_REDIRECT_URI;

export async function GET() {
  const scope =
    "user-read-private user-read-email user-read-currently-playing user-read-playback-state";
  const state = "some-random-state"; // Ideally, generate a random state for security

  const authUrl =
    "https://accounts.spotify.com/authorize?" +
    querystring.stringify({
      response_type: "code",
      client_id: client_id,
      scope: scope,
      redirect_uri: redirect_uri,
      state: state,
    });

  return NextResponse.redirect(authUrl);
}
