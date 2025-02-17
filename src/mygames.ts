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