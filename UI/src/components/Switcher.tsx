import type { Dispatch } from "react";

const SwitchButton = ({
	isOn,
	setIsOn,
}: {
	isOn: boolean;
	setIsOn: Dispatch<boolean>;
}) => {
	// Inline styles for the button container and circle
	const buttonStyle = {
		display: "flex",
		alignItems: "center",
		justifyContent: isOn ? "flex-end" : "flex-start",
		width: "60px",
		height: "30px",
		backgroundColor: isOn ? "#4caf50" : "#ccc",
		borderRadius: "30px",
		cursor: "pointer",
		padding: "5px",
		transition: "background-color 0.3s, justify-content 0.3s",
	};

	const circleStyle = {
		width: "20px",
		height: "20px",
		backgroundColor: "white",
		borderRadius: "50%",
		boxShadow: "0 0 2px rgba(0,0,0,0.3)",
	};

	const labelStyle = {
		marginLeft: "10px",
		fontFamily: "Arial, sans-serif",
		fontSize: "14px",
		color: isOn ? "#4caf50" : "#555",
	};

	return (
		<div style={{ display: "flex", alignItems: "center" }}>
			<div style={buttonStyle} onClick={() => setIsOn(!isOn)}>
				<div style={circleStyle}></div>
			</div>
			<span style={labelStyle}>{isOn ? "Download" : "Watch"}</span>
		</div>
	);
};

export default SwitchButton;
