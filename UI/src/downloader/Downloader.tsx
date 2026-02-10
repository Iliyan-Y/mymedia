import { useState } from "react";

const baseUrl = "/api";
const baseMediaUrl = "/api";

export default function Downloader() {
	const [url, setUrl] = useState("");
	const [progress, setProgress] = useState("");
	const [status, setStatus] = useState("");
	const [downloading, setDownloading] = useState(false);
	const [filename, setFilename] = useState<string | null>();

	const handleDownload = async () => {
		setDownloading(true);
		setProgress("");
		setStatus("Starting...");

		try {
			const res = await fetch(
				`${baseUrl}/download?url=${encodeURIComponent(url)}`,
			);
			const data: { job_id: string } = await res.json();
			listenToProgress(data.job_id);
		} catch (err) {
			setStatus("Error starting download");
			console.error(err);
			setDownloading(false);
		}
	};

	const listenToProgress = (id: string) => {
		const evt = new EventSource(`${baseUrl}/progress/stream?id=${id}`);
		evt.onmessage = (e) => {
			const data = JSON.parse(e.data);
			setProgress(data.progress);

			if (data.filename && data.filename.length > 0) {
				setFilename(data.filename);
			}

			if (data.progress.includes("done") || data.progress.includes("100%")) {
				setStatus("✅ Download complete");
				evt.close();
				setDownloading(false);
			}
		};

		evt.onerror = (e) => {
			console.error("SSE error:", e);
			evt.close();
			setStatus("⚠️ Connection lost");
			setDownloading(false);
		};
	};

	return (
		<div style={{ width: "100%" }}>
			<h2>🎥 Video Downloader</h2>

			<input
				type="text"
				value={url}
				onChange={(e) => setUrl(e.target.value)}
				placeholder="Enter video URL..."
				style={{ width: "100%", padding: "8px" }}
			/>

			<button
				onClick={handleDownload}
				disabled={!url || downloading}
				style={{ marginTop: "1rem", padding: "10px 20px", cursor: "pointer" }}
			>
				{downloading ? "Downloading..." : "Start Download"}
			</button>

			{status && <p style={{ marginTop: "1rem" }}>{status}</p>}

			{progress && (
				<pre
					style={{
						background: "#222",
						color: "#0f0",
						padding: "1rem",
						marginTop: "1rem",
						overflowY: "auto",
					}}
				>
					{progress}
					<br />
					{filename}
				</pre>
			)}

			{filename && (
				<video
					src={`${baseMediaUrl}/api/media/stream/${filename}`}
					controls
					autoPlay={false}
					width="340"
				/>
			)}
		</div>
	);
}
