import YouTube from "react-youtube";
import { useEffect, useRef, useState } from "react";

export default function YouTubePlayer({
  videoId,
  playing,
  onPause,
  onPlayerReady,
}) {
  const playerRef = useRef(null);
  const [ready, setReady] = useState(false);

  const opts = {
    width: "0", // hide width
    height: "0", // hide height
    playerVars: {
      modestbranding: 1,
      rel: 0,
      // Specify origin so that the YouTube iframe matches the current host
      origin: window.location.origin,
    },
  };

  const onReady = (event) => {
    playerRef.current = event.target;
    setReady(true);
    if (onPlayerReady) {
      onPlayerReady();
    }
  };

  useEffect(() => {
    setReady(false);
    playerRef.current = null;
  }, [videoId]);

  useEffect(() => {
    if (ready) {
      if (playing) {
        playerRef.current.playVideo();
      } else {
        playerRef.current.pauseVideo();
      }
    }
  }, [playing, ready]);

  return (
    <YouTube
      videoId={videoId}
      opts={opts}
      onReady={onReady}
      style={{ display: "none" }}
    />
  );
}
