import { useEffect, useState } from "react";

const baseMediaUrl = "http://localhost:8081";

const MediaPlayer = () => {
	const [files, setFiles] = useState([]);
	const [current, setCurrent] = useState(null);

	useEffect(() => {
		fetch(`${baseMediaUrl}/api/media`)
			.then((res) => res.json())
			.then(setFiles);
	}, []);

	return (
		<div>
			<h3>Media Player</h3>
			<div>
				{files.map((f) => (
					<div key={f}>
						<button onClick={() => setCurrent(f)}>{f}</button>
					</div>
				))}
			</div>

			{current && (
				<video
					src={`${baseMediaUrl}/api/media/stream/${current}`}
					controls
					autoPlay
					width="640"
				/>
			)}
		</div>
	);
};

export default MediaPlayer;
