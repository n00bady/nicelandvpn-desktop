import { Navigate } from "react-router-dom";
import React, { useState } from "react";
import STORE from "../store";
import { DesktopIcon, MagnifyingGlassIcon, EnterIcon } from "@radix-ui/react-icons";
import API from "../api";

const Routers = (props) => {

	const [filter, setFilter] = useState("");

	const switchRouter = async (router) => {


		if (router.Tag === "") {
			props.toggleLoading({ tag: "ROUTERS", show: true, msg: "Switching to automatic router selection" })
		} else {
			props.toggleLoading({ tag: "ROUTERS", show: true, msg: "Switching to " + router.Tag })
		}

		let x = await API.method("switchRouter", { Tag: router.Tag })
		if (x === undefined) {
			props.toggleError("Unknown error, please try again in a moment");

		} else {
			if (x.status === 200) {
				props.showSuccessToast("Router switch complete")
			} else {
				props.toggleError("Unknown error, please try again in a moment");
			}
		}

		props.toggleLoading(undefined)
	}


	let routers = []

	if (props?.state?.Routers) {

		if (filter && filter !== "") {
			props.state.Routers.map(r => {

				let filterMatch = false
				if (r.Tag.includes(filter)) {
					filterMatch = true
				}
				if (r.Country.includes(filter)) {
					filterMatch = true
				}

				if (filterMatch) {
					routers.push(r)
				}
			})

		} else {
			routers = props.state.Routers
		}

	}
	console.log("ROUTER COUNT:", routers.length)
	console.dir(routers)

	const RenderServer = (s, active) => {

		let country = "icon"
		if (s.Country !== "") {
			country = s.Country.toLowerCase()
		}

		return (
			<div className={`router ${active ? "active-bg" : ""}`} key={s.Tag} onClick={() => switchRouter(s)}>
				<div className="item index">
					<span className="green">{s.ListIndex}</span>
				</div>

				{active &&
					<div className="item active-icon">
						<EnterIcon
							height={23}
							width={23}
							className="">
						</EnterIcon>
					</div>
				}

				{s.Tag &&
					<div className="item tag">
						{s.Tag}
					</div>
				}

				{!s.Tag &&
					<div className="item ip">Unknown</div>
				}

				<div className="item country" >
					{country !== "icon" &&
						<>
							<img
								className="country-flag"
								src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/" + country + ".svg"}
							/>
							<div className="text">
								{country.toUpperCase()}
							</div>
						</>
					}
					{country === "icon" &&
						<>
							<DesktopIcon className="country-temp" height={23} width={23}></DesktopIcon>
							<div className="text green">
								Private
							</div>
						</>
					}

				</div>

				<div className="item slots">
					Slots
					<span className="green">{s.Slots}</span>
				</div>
				<div className="item ">
					Score
					<span className="green">{s.Score}</span>
				</div>
				<div className="item ">
					MS
					<span className="green">{s.MS}</span>
				</div>
				<div className="item ">
					UserMbps
					<span className="green">{s.AvailableUserMbps}</span>
				</div>
				<div className="item ">
					Mbps
					<span className="green">{s.AvailableMbps}</span>
				</div>
			</div>
		)
	}

	let AR = props.state?.ActiveRouter

	return (
		<div className="router-wrapper-new"  >

			<div className="search-wrapper">
				<MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
				<input type="text" className="search" onChange={(e) => setFilter(e.target.value)} placeholder="Search .."></input>
			</div>

			{props.state?.C?.ManualRouter &&
				<div className="automatic-button"
					onClick={() => switchRouter({ Tag: "" })} >Switch Back To Automatic Router Selection</div>
			}

			<div className="routers">

				{routers.map(r => {
					if (AR) {
						if (AR.Tag === r.Tag) {
							return RenderServer(r, true)
						} else {
							return RenderServer(r, false)
						}
					} else {
						return RenderServer(r, false)
					}
				})}

			</div>

		</div >
	);
}

export default Routers;
