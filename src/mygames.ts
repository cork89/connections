import { isErrorWithProperty } from "./common.js"

const copyNotif = document.getElementsByClassName("copynotif")[0]
const copyEls = document.getElementsByTagName("img")

for (let i = 0; i < copyEls.length; i++) {
    const gameLink = copyEls[i]

    if (gameLink) {
        const grandparent = gameLink?.parentElement?.parentElement
        const anchortag = grandparent?.querySelector("a")

        if (anchortag) {
            copyEls[i]?.addEventListener("click", () => copyGamelink(anchortag))
        }
    }

}

async function copyGamelink(gameLink: HTMLAnchorElement) {
    try {
        await navigator.clipboard.writeText(gameLink.href)

        if (copyNotif) {
            (copyNotif as HTMLElement).style.visibility = "visible"
            copyNotif.innerHTML = `Copied <b>${gameLink.innerHTML}</b> to clipboard!`

            setTimeout(() => {
                (copyNotif as HTMLElement).style.visibility = "hidden"
            }, 3000)
        }
    } catch (error: unknown) {
        if (isErrorWithProperty(error, "message")) {
            console.error(error.message)
        }
    }
}

var tpCheckbox = document.getElementById("tp-checkbox") as HTMLElement
var myGames = document.getElementById("my-games") as HTMLElement
var recentGames = document.getElementById("recent-games") as HTMLElement
var url = new URL(window.location.toString())
var urlParams = new URLSearchParams(url.search)

// Flip from the created to recent played tables and vice versa
function flipTables(): void {
    if (myGames.classList.contains("void")) {
        myGames.classList.remove("void")
        recentGames.classList.add("void")
        urlParams.set("table", "created")
    } else {
        myGames.classList.add("void")
        recentGames.classList.remove("void")
        urlParams.set("table", "recent")
    }
    url.search = urlParams.toString()
    window.history.pushState(null, "", url.toString())
}

tpCheckbox.addEventListener("change", flipTables)