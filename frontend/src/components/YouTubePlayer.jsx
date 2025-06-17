import YouTube from "react-youtube";
import { useEffect, useRef, useState } from "react";

export default function YouTubePlayer({ videoId, playing, onPause }) {
  const playerRef = useRef(null);
  const [ready, setReady] = useState(false);

  const opts = {
    width: "640",
    height: "360",
    playerVars: {
      modestbranding: 1,
      rel: 0,
    },
  };

  const onReady = (event) => {
    playerRef.current = event.target;
    setReady(true);
  };

  useEffect(() => {
    if (ready) {
      if (playing) {
        playerRef.current.playVideo();
      } else {
        playerRef.current.pauseVideo();
      }
    }
  }, [playing, ready]);

  return <YouTube videoId={videoId} opts={opts} onReady={onReady} />;
}
