import { useEffect, useState } from "react";

const baseMediaUrl = "";

const MediaPlayer = () => {
	const [files, setFiles] = useState([]);
	const [current, setCurrent] = useState(null);

	useEffect(() => {
		fetch(`${baseMediaUrl}/api/media`)
			.then((res) => res.json())
			.then(setFiles);
	}, []);

	if (!files || files.length === 0) return <div>NoFiles...</div>;

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
