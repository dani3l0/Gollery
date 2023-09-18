document.addEventListener("DOMContentLoaded", function() {
	try {
		// Image Viewer
		let image = document.getElementById("image")
		image.addEventListener("load", function() {
			image.classList.add("loaded")
		})
		let viewer = document.getElementById("viewer")
		viewer.addEventListener("click", function() {
			viewer.classList.remove("open")
			image.classList.remove("loaded")
		})

		// Image list
		let imageItems = document.getElementById("images").children
		if (!imageItems.length) {
			document.getElementById("images").parentNode.style.display = "none"
		}
		for (let item of imageItems) {
			item.addEventListener("click", function(e) {
				e.preventDefault()
				showImage(item)
			})
		}

		// Folder list
		let folderItems = document.getElementById("folders").children
		if (!folderItems.length) {
			document.getElementById("folders").parentNode.style.display = "none"
		}
	} catch(e) {}
})

function set(id, value) {
	document.getElementById(id).innerText = value
}

function settings() {
	settingsMain()
	setInterval(settingsMain, 2000)
}

function settingsMain() {
	let xhr = new XMLHttpRequest()
	xhr.open("GET", "/settings/api")
	xhr.onload = function() {
		if (this.status == 200) {
			let data = JSON.parse(this.responseText)
			set("images", data.Images)
			set("storage", `${data.Size.toFixed(3)} GB`)
			set("cache", `${data.Cache.toFixed(3)} GB`)
			set("free", `${data.Free.toFixed(3)} GB`)

			if (!data.Scanning) {
				let lastScan = data.LastScan
				if (lastScan) {
					let d = new Date(lastScan)
					const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
					lastScan = `${d.getDate()} ${months[d.getMonth()]} at ${d.getHours()}:${d.getMinutes()}`
				}
				else {
					lastScan = "Long time ago"
				}
				set("last-scan", lastScan)
			}
			else {
				set("last-scan", "Scanning now")
			}
		}
	}
	xhr.send()
}

function scan() {
	let xhr = new XMLHttpRequest()
	xhr.open("GET", "/settings/scan")
	xhr.send()
}

function stopScan() {
	let xhr = new XMLHttpRequest()
	xhr.open("GET", "/settings/scanStop")
	xhr.send()
}

function showImage(img) {
	document.getElementById("viewer").classList.add("open")
	document.getElementById("image").src = img.href
}
