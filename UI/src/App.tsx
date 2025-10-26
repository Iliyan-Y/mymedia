import { useState } from "react";
import "./App.css";
import Downloader from "./downloader/Downloader";
import MediaPlayer from "./MediaPlayer/MediaPlayer";
import Switcher from "./components/Switcher";

function App() {
	const [showDLPage, setShowDLPage] = useState(false);

	return (
		<div
			style={{ maxWidth: 600, margin: "2rem auto", fontFamily: "sans-serif" }}
		>
			<Switcher isOn={showDLPage} setIsOn={setShowDLPage} />

			{showDLPage ? <Downloader /> : <MediaPlayer />}
		</div>
	);
}

export default App;
