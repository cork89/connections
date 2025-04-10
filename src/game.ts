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
    selected: Array<WordSelection>
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
    const request: SelectionRequest = { selected: [] }

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
            method: "POST"
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
}

setupGame()