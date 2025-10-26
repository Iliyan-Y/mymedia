import { useState } from "react";

const baseUrl = "http://localhost:8080";

export default function Downloader() {
	const [url, setUrl] = useState("");
	//	const [jobId, setJobId] = useState("");
	const [progress, setProgress] = useState("");
	const [status, setStatus] = useState("");
	const [downloading, setDownloading] = useState(false);

	const handleDownload = async () => {
		setDownloading(true);
		setProgress("");
		setStatus("Starting...");

		try {
			const res = await fetch(
				`${baseUrl}/download?url=${encodeURIComponent(url)}`
			);
			const data: { job_id: string } = await res.json();
			//setJobId(data.job_id);
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
			const text = e.data;
			setProgress(text);
			if (text.includes("100%") || text.includes("done")) {
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
		<div
			style={{ maxWidth: 600, margin: "2rem auto", fontFamily: "sans-serif" }}
		>
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
						height: "200px",
						overflowY: "auto",
					}}
				>
					{progress}
				</pre>
			)}
		</div>
	);
}
