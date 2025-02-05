/**
 * An individual word selection.
 * @typedef {Object} Selection
 * @property {number} id
 * @property {string} word
 */

/**
 * The selection request.
 * @typedef {Object} SelectionRequest
 * @property {Selection[]} selected the array of selected words.
 */

const words = document.getElementsByClassName("word")
var guesses = []
var currentlySelected = new Set()
var nextBoard = ""

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
        window.removeEventListener("click", closeModal2)
        guesses.splice(0, guesses.length)
        init()
    } catch (error) {
        console.error(error.message);
    }
}

/**
 * Creates the request of selected words.
 * @returns {SelectionRequest}
 */
function getSelectionRequest() {
    const request = { selected: [] }

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
            checkbutton.disabled = true
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
            let close = modal.children[0].children[0]
            close.addEventListener("click", closeModal)
            window.addEventListener("click", closeModal2)
        }

        const oneaway_p = oneaway.children[0]
        if (!oneaway_p.classList.contains("hidden")) {
            setTimeout(() => {
                oneaway_p.classList.add("hidden")
            }, 3000)
        }

        init()

    } catch (error) {
        console.error(error.message);
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
        console.error(error.message);
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
    init()
}

/**
 * Updates all of the currently selected cursors to pointer. 
 */
function resetPointers() {
    for (let i = 0; i < words.length; i++) {
        if (!currentlySelected.has(words[i].id)) {
            words[i].style.cursor = "pointer"
        }
    }
}

/**
 * Deselect all of the words.
 */
async function deselectAll() {
    // console.log(currentlySelected)
    for (const value of currentlySelected) {
        let el = document.getElementById(value)
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
        console.error(error.message);
        return
    }
}

/**
 * Check is two sets are equal.
 * @param {Set<string>} set1
 * @param {Set<string>} set2
 * @return {boolean} true if sets are equal, false otherwise
 */
function areSetsEqual(set1, set2) {
    if (set1.size !== set2.size) {
        return false;
    }

    for (const element of set1) {
        if (!set2.has(element)) {
            return false;
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
function containsSet(array1, set2) {
    let contained = false
    for (let i = 0; i < array1.length; i++) {
        if (areSetsEqual(array1[i], set2)) {
            contained = true
        }
    }
    return contained
}

/**
 * Attempt select a word. Only 4 words can be selected at a time. If a word is already selected, this should unselect it.
 * @param {PointerEvent} event the click event on a word
 */
function attemptSelection(event) {
    let el;
    if (event.target.nodeName == "SPAN") {
        el = event.target.parentElement
    } else {
        el = event.target
    }
    let elExists = currentlySelected.has(el.id)
    checkbutton.disabled = true
    if (elExists) {
        el.classList.remove("selected")
        currentlySelected.delete(el.id)
    } else if (currentlySelected.size < 4) {
        el.classList.add("selected")
        currentlySelected.add(el.id)
    }

    const status = document.getElementsByClassName("gamestatus")[0].innerText
    // console.log(guesses)
    if (currentlySelected.size == 4 && !containsSet(guesses, currentlySelected) && status != "loser") {
        checkbutton.disabled = false
    }

    if (currentlySelected.size == 4) {
        for (let i = 0; i < words.length; i++) {
            if (!currentlySelected.has(words[i].id)) {
                words[i].style.cursor = "default"
            }
        }
    } else {
        resetPointers()
    }
}


/**
 * When the user clicks on <span> (x), close the modal
 */
function closeModal() {
    const modal = document.getElementById("gameOverModal")
    modal.style.display = "none"
}

/**
 * When the user clicks anywhere outside of the modal, close it
 * @param {PointerEvent} event 
 */
function closeModal2(event) {
    const modal = document.getElementById("gameOverModal")
    if (event.target == modal) {
        modal.style.display = "none";
    }
}

/**
 * Initialize the board by adding a click handler to each word for attempting selection.  Reset the selected words.
 */
function init() {
    for (let i = 0; i < words.length; i++) {
        words[i].addEventListener("click", attemptSelection)
    }

    const selected = document.getElementsByClassName("selected")
    currentlySelected.clear()
    for (let i = 0; i < selected.length; i++) {
        currentlySelected.add(selected[i].id)
    }
}

init()