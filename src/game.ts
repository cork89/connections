import { isErrorWithProperty } from "./common.js"

/**
 * An individual word selection.
 */
type WordSelection = {
    id: number,
    word: string
}

/**
 * The selection request.
 */
type SelectionRequest = {
    selected: Array<WordSelection>,
    hintsRevealed: boolean
}

const words = document.getElementsByClassName("word")
var guesses: Array<any> = []
var currentlySelected: Set<any> = new Set()
var nextBoard = ""

const board: HTMLElement = document.getElementById("board") ?? (() => { throw new Error("board cannot be null") })()
const shuffleButton: HTMLButtonElement = document.getElementById("shuffleButton") as HTMLButtonElement ?? (() => { throw new Error("shuffleButton cannot be null") })()
const deselectButton: HTMLButtonElement = document.getElementById("deselectButton") as HTMLButtonElement ?? (() => { throw new Error("deselectButton cannot be null") })()
const checkButton: HTMLButtonElement = document.getElementById("checkButton") as HTMLButtonElement ?? (() => { throw new Error("checkButton cannot be null") })()
const resetButton: HTMLButtonElement | null = document.getElementById("resetButton") as HTMLButtonElement | null
// const oneaway: HTMLElement = document.getElementById("oa") ?? (() => { throw new Error("oneaway cannot be null") })()
const closeModalButton: HTMLElement | null = document.getElementById("close-modal")
const closeModalWindow: HTMLElement | null = document.getElementById("gameOverModal")

shuffleButton.addEventListener("click", shuffle)
deselectButton.addEventListener("click", deselectAll)
checkButton.addEventListener("click", check)

if (resetButton) {
    resetButton.addEventListener("click", reset)
}
if (closeModalWindow && closeModalButton) {
    closeModalButton.addEventListener("click", () => closeModal(closeModalWindow))
}

if (closeModalWindow) {
    closeModalWindow.addEventListener("click", (event) => closeModalEmptySpace(event as PointerEvent))
}

/**
 * Resets the game board to the initial state. 
 */
async function reset() {
    const url = "reset/"
    try {
        const response = await fetch(url, {
            headers: {
                "Accept": "text/html"
            },
            method: "POST"
        })
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`)
        }

        const text = await response.text()
        if (text == "nochange") {
            return
        }
        board.innerHTML = text
        const modal = document.getElementById("gameOverModal")
        if (modal) {
            modal.removeEventListener("click", (event) => closeModalEmptySpace(event as PointerEvent))
        }

        guesses.splice(0, guesses.length)
        setupGame()
    } catch (error) {
        if (isErrorWithProperty(error, "message")) {
            console.error(error.message)
        };
    }
}

/**
 * Creates the request of selected words.
 * @returns {SelectionRequest}
 */
function getSelectionRequest() {
    const request: SelectionRequest = { selected: [], hintsRevealed: (hintsStage === HintStage.REVEALED_INVIS || hintsStage === HintStage.REVEALED_VIS) ? true : false }

    for (const value of currentlySelected) {
        request.selected.push({
            id: parseInt(value.slice(4,)),
            word: value
        })
    }
    return request
}

/**
 * Checks the selected words to see if they match a category.
 */
async function check() {
    const url = "check/"
    try {
        const request = getSelectionRequest()

        if (currentlySelected.size == 4) {
            guesses.push(new Set(currentlySelected))
            checkButton.disabled = true
        }

        const response = await fetch(url, {
            headers: {
                "Accept": "text/html",
                "Content-Type": "application/json"
            },
            method: "POST",
            body: JSON.stringify(request)
        })
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`)
        }

        const text = await response.text()
        if (text == "nochange") {
            return
        }
        board.innerHTML = text

        const modal = document.getElementById("gameOverModal")

        if (modal) {
            const close = modal.children[0].children[0]
            close.addEventListener("click", () => closeModal(modal))
            modal.addEventListener("click", (event) => closeModalEmptySpace(event as PointerEvent));
        }

        const oneaway_span = document.getElementById("oa")?.children[0]
        if (oneaway_span && !oneaway_span.classList.contains("hidden")) {
            setTimeout(() => {
                oneaway_span.classList.add("hidden")
            }, 3000)
        }

        setupGame()

    } catch (error) {
        if (isErrorWithProperty(error, "message")) {
            console.error(error.message)
        }
    }
}

/**
 * Calls the shuffle endpoint.
 * @returns {string} the new board html text
 */
async function shuffleData() {
    try {
        const url = "shuffle/"
        const request = getSelectionRequest()
        const response = await fetch(url, {
            headers: {
                "Accept": "text/html",
                "Content-Type": "application/json"
            },
            method: "POST",
            body: JSON.stringify(request)
        })
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`)
        }
        return await response.text()
    } catch (error) {
        if (isErrorWithProperty(error, "message")) {
            console.error(error.message)
        };
        return ""
    }
}

/**
 * Shuffles the board.
 */
async function shuffle() {
    // if (nextBoard == "") {
    //     nextBoard = await shuffleData()
    // }
    // if (nextBoard != "") {
    //     board.innerHTML = nextBoard
    // }
    board.innerHTML = await shuffleData()
    // nextBoard = await shuffleData()
    setupGame()
}

/**
 * Updates all of the currently selected cursors to pointer. 
 */
function resetPointers() {
    for (let i = 0; i < words.length; i++) {
        if (!currentlySelected.has(words[i].id)) {
            (words[i] as HTMLElement).style.cursor = "pointer"
        }
    }
}

/**
 * Deselect all of the words.
 */
async function deselectAll() {
    for (const value of currentlySelected) {
        let el = document.getElementById(value) ?? (() => { throw new Error(`${value} cannot be null`) })()
        el.classList.remove("selected")
    }

    currentlySelected.clear()
    // nextBoard = ""
    resetPointers()
    // console.log(currentlySelected)
    try {
        const url = "deselectAll/"
        const response = await fetch(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                "hintsRevealed": (hintsStage === HintStage.REVEALED_INVIS || hintsStage === HintStage.REVEALED_VIS) ? true : false
            })
        })
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`)
        }
    } catch (error) {
        if (isErrorWithProperty(error, "message")) {
            console.error(error.message)
        }
        return
    }
}

/**
 * Check is two sets are equal.
 * @param {Set<string>} set1
 * @param {Set<string>} set2
 * @return {boolean} true if sets are equal, false otherwise
 */
function areSetsEqual(set1: Set<string>, set2: Set<string>): boolean {
    if (set1.size !== set2.size) {
        return false;
    }

    for (const element of set1) {
        if (!set2.has(element)) {
            return false
        }
    }

    return true;
}

/**
 * Check if an array of Sets contains a given set.
 * @param {Set<string>[]} array1
 * @param {Set<string>} set2
 * @return {boolean} true if array contains equivalent set, false otherwise
 */
function containsSet(array1: Array<Set<string>>, set2: Set<string>): boolean {
    let contained = false
    for (let i = 0; i < array1.length; i++) {
        if (areSetsEqual(array1[i], set2)) {
            contained = true
            break
        }
    }
    return contained
}

/**
 * Attempt select a word. Only 4 words can be selected at a time. If a word is already selected, this should unselect it.
 * @param {PointerEvent} event the click event on a word
 */
function attemptSelection(event: any) {
    let el;
    if (event.target.nodeName == "SPAN") {
        el = event.target.parentElement
    } else {
        el = event.target
    }
    let elExists = currentlySelected.has(el.id)
    checkButton.disabled = true
    if (elExists) {
        el.classList.remove("selected")
        currentlySelected.delete(el.id)
    } else if (currentlySelected.size < 4) {
        el.classList.add("selected")
        currentlySelected.add(el.id)
    }

    const status: HTMLCollectionOf<Element> = document.getElementsByClassName("gamestatus") ?? (() => { throw new Error("gamestatus cannot be null") })()
    const statusText = (status[0] as HTMLElement).innerText
    // console.log(guesses)
    if (currentlySelected.size == 4 && !containsSet(guesses, currentlySelected) && statusText != "loser") {
        checkButton.disabled = false
    }

    if (currentlySelected.size == 4) {
        for (let i = 0; i < words.length; i++) {
            if (!currentlySelected.has(words[i].id)) {
                (words[i] as HTMLElement).style.cursor = "default"
            }
        }
    } else {
        resetPointers()
    }
}


/**
 * When the user clicks on <span> (x), close the modal
 */
function closeModal(el: HTMLElement) {
    el.style.display = "none"
}

/**
 * When the user clicks anywhere outside of the modal, close it
 * @param {PointerEvent} event 
 */
function closeModalEmptySpace(event: PointerEvent) {
    if (closeModalWindow && event.target === closeModalWindow) {
        closeModalWindow.style.display = "none";
    }
}

enum HintStage {
    NOT_REVEALED_INVIS = "notrevealedinvis",
    NOT_REVEALED_VIS = "notrevealedvis",
    REVEALED_INVIS = "revealedinvis",
    REVEALED_VIS = "revealedvis",
}

const revealHint: HTMLElement = document.getElementById("reveal-hint") as HTMLElement
var hintsStage: HintStage = (revealHint.innerText === "Hints:") ? HintStage.REVEALED_INVIS : HintStage.NOT_REVEALED_INVIS
var hintInterval: number = 0
var hintTimer: number = 5

// Handler for showing hints, only show hints after a 5 second timeout
function handleHintPress() {
    const hintContainer: HTMLElement = document.getElementsByClassName("hint-container")[0] as HTMLElement
    const revealHint: HTMLElement = document.getElementById("reveal-hint") as HTMLElement
    const yellowHint: HTMLElement = document.getElementById("yellow-hint") as HTMLElement
    const greenHint: HTMLElement = document.getElementById("green-hint") as HTMLElement
    const blueHint: HTMLElement = document.getElementById("blue-hint") as HTMLElement
    const purpleHint: HTMLElement = document.getElementById("purple-hint") as HTMLElement


    if (hintsStage === HintStage.NOT_REVEALED_INVIS) {
        hintContainer.classList.add("open")
        hintsStage = HintStage.NOT_REVEALED_VIS
        hintInterval = setInterval(function () {
            revealHint.innerText = `Hints revealing in... ${--hintTimer}`
            if (hintTimer <= 0) {
                revealHint.innerText = "Hints:";
                (yellowHint.childNodes[0] as HTMLElement).textContent = "🟡 ";
                (yellowHint.childNodes[1] as HTMLElement).style = "";
                (greenHint.childNodes[0] as HTMLElement).textContent = "🟢 ";
                (greenHint.childNodes[1] as HTMLElement).style = "";
                (blueHint.childNodes[0] as HTMLElement).textContent = "🔵 ";
                (blueHint.childNodes[1] as HTMLElement).style = "";
                (purpleHint.childNodes[0] as HTMLElement).textContent = "🟣 ";
                (purpleHint.childNodes[1] as HTMLElement).style = "";
                hintsStage = HintStage.REVEALED_VIS
                clearInterval(hintInterval)
            }
        }, 1000)
    } else if (hintsStage === HintStage.REVEALED_INVIS) {
        hintContainer.classList.add("open")
        hintsStage = HintStage.REVEALED_VIS
    }
}

// Handler for hiding hints
function handleHintRelease() {
    const hintContainer: HTMLElement = document.getElementsByClassName("hint-container")[0] as HTMLElement
    // hintContainer.classList.remove("open")
    if (hintsStage === HintStage.NOT_REVEALED_VIS) {
        hintContainer.classList.remove("open")
        clearInterval(hintInterval)
        hintsStage = HintStage.NOT_REVEALED_INVIS
    } else if (hintsStage === HintStage.REVEALED_VIS) {
        hintContainer.classList.remove("open")
        hintsStage = HintStage.REVEALED_INVIS
    }
}

/**
 * Initialize the board by adding a click handler to each word for attempting selection.  Reset the selected words.
 */
function setupGame() {
    for (let i = 0; i < words.length; i++) {
        words[i].addEventListener("click", attemptSelection)
    }

    const selected = document.getElementsByClassName("selected")
    currentlySelected.clear()
    for (let i = 0; i < selected.length; i++) {
        currentlySelected.add(selected[i].id)
    }
    const hintIcon = document.getElementById("hint-icon") as HTMLElement
    hintIcon.addEventListener("mousedown", handleHintPress)
    hintIcon.addEventListener("mouseup", handleHintRelease)
    hintIcon.addEventListener("touchstart", handleHintPress)
    hintIcon.addEventListener("touchend", handleHintRelease)
}

setupGame()