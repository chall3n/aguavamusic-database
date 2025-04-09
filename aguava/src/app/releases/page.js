"use client";

import Image from "next/image";
import { useEffect, useState } from "react";
import Link from "next/link";
import "./style.css";

export default function Releases() {
  const [songs, setSongs] = useState([]);
  const [sortBy, setSortBy] = useState("streams");
  const [sortOrder, setSortOrder] = useState("desc");

  useEffect(() => {
    const fetchSongs = async () => {
      try {
        const response = await fetch(
          `http://localhost:8080/songs?sortBy=${sortBy}&sortOrder=${sortOrder}`
        );
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        setSongs(data);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };
    fetchSongs();
  }, [sortBy, sortOrder]);

  const handSortChange = (event) => {
    setSortBy(event.target.value);
  };

  const handleSortOrderToggle = () => {
    setSortOrder((prevOrder) => (prevOrder === "asc" ? "desc" : "asc"));
  };

  return (
    <div>
      <div className="sort-container">
        <button className="sort-button" onClick={handleSortOrderToggle}>
          Sort Order: {sortOrder === "asc" ? "Ascending" : "Descending"}
        </button>
        <select
          id="sortOptions"
          className="sort-dropdown"
          value={sortBy}
          onChange={handSortChange}
        >
          <option value="streams">Streams</option>
          <option value="key">Key</option>
          <option value="bpm">bpm</option>
          <option value="popularity">Popularity</option>
        </select>
      </div>

      <h2>Aguava</h2>
      <Link href="/">Back to Home</Link>
      {songs.map(
        (song) => (
          console.log("Rendering song:", song),
          console.log("Spotify URL:", song.spot),
          (
            <div key={song.name} className="song-card">
              <div className="song-detail">
                <span>"{song.name}"</span>
                <span>{song.streams} streams</span>
                <span>Key of {song.key}</span>
                <span>{song.bpm} bpm</span>
                <span>Popularity: {song.popularity}</span>
                <div>
                  <div>
                    {song.spot ? (
                      <iframe
                        width="300"
                        height="80"
                        src={song.spot} // Ensure this contains the correct Spotify URL
                        frameBorder="0"
                        allow="encrypted-media"
                        allowtransparency="true"
                      ></iframe>
                    ) : (
                      <p>No Spotify embed available</p>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )
        )
      )}
    </div>
  );
}
