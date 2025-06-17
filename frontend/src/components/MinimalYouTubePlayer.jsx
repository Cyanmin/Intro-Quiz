import { useEffect, useRef, useState } from "react";
import YouTube from "react-youtube";

// Minimal player component that waits for the iframe to be ready
// before sending play or pause commands.
export default function MinimalYouTubePlayer({ videoId, playing }) {
  const playerRef = useRef(null);
  const [isReady, setIsReady] = useState(false);

  const opts = {
    playerVars: {
      // Setting origin and enablejsapi is required for postMessage to work
      origin: window.location.origin,
      enablejsapi: 1,
    },
  };

  const handleReady = (event) => {
    // Store the player instance and mark as ready
    playerRef.current = event.target;
    setIsReady(true);
  };

  useEffect(() => {
    if (!isReady) return;

    if (playing) {
      playerRef.current.playVideo();
    } else {
      playerRef.current.pauseVideo();
    }
  }, [playing, isReady]);

  return (
    <YouTube videoId={videoId} opts={opts} onReady={handleReady} />
  );
}
