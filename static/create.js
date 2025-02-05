/**
 * Contains the possible states for creating a connections game.
 */
const States = Object.freeze({
    NONE: "none",
    YELLOW: "yellow",
    GREEN: "green",
    BLUE: "blue",
    PURPLE: "purple",
})

/**
 * Represents a single category with the words associated to that category.
 * @typedef {Object} ColorCategory
 * @param {string} category
 * @param {Array<string>} words
 */

/**
 * Represents the four connections categories and words related to each category.
 * @typedef {Object} Categories
 * @property {ColorCategory} yellow
 * @property {ColorCategory} green
 * @property {ColorCategory} blue
 * @property {ColorCategory} purple
 */

/**
 * Represents the context of the page storing the state, categories, and custom game id.
 * @typedef {Object} Context
 * @property {States} state
 * @property {Categories} categories
 * @property {string} gameId
 */

/**
 * Creates an empty object for storing the state of each connection color.
 * @returns {Categories}
 */
function getEmptyCategories() {
    return {
        "yellow": {
            "category": "",
            "words": []
        },
        "green": {
            "category": "",
            "words": []
        },
        "blue": {
            "category": "",
            "words": []
        },
        "purple": {
            "category": "",
            "words": []
        }
    }
}

/**
 * Creates a default context object
 * @returns {Context}
 */
function getDefaultContext() {
    return {
        "state": States.NONE,
        "categories": getEmptyCategories(),
        "gameId": ""
    }
}

/**
 * Stores meta data for the users created games in local storage
 * @param {string} gameId 
 * @param {number} timestamp 
 * @param {Array<string>} categories 
 */
function storeGame(gameId, timestamp, categories) {
    const game = {
        "gameId": gameId,
        "timestamp": timestamp,
        "categories": categories,
    }
    createdGames.push(game)
    localStorage.setItem("createdGames", JSON.stringify(createdGames))
}

let ctx = JSON.parse(localStorage.getItem("ctx"))
if (!ctx) {
    ctx = getDefaultContext()
    localStorage.setItem("ctx", JSON.stringify(ctx))
}

let createdGames = JSON.parse(localStorage.getItem("createdGames"))
if (!createdGames) {
    createdGames = []
    localStorage.setItem("createdGames", JSON.stringify(createdGames))
}

const categoryInput = document.getElementById("category-input")
const wordsInput = document.getElementById("word-input")
const saveCategoriesButton = document.getElementById("save-categories-button")
const submitButton = document.getElementById("submit")

var currentWords = []
var currentCategory = []

// /**
//  * Get the 
//  * @param {string} color 
//  * @returns 
//  */
// function statePicker(color) {
//     if (color == "yellow") {
//         return States.YELLOW
//     } else if (color == "blue") {
//         return States.BLUE
//     } else if (color == "green") {
//         return States.GREEN
//     } else if (color == "purple") {
//         return States.PURPLE
//     }
//     console.error("bad color")
// }

/**
 * Displays all of the categories and words for each color group from context.
 */
function displayCategoryAndWords() {
    for (const key of Object.keys(ctx.categories)) {
        // console.log(key, ctx.categories[key])
        const colorCategory = document.getElementById(`category-${key}`)
        const colorWords = document.getElementById(`words-${key}`)
        let newCategoryHTML = "tbd"
        if (ctx.categories[key].category != "") {
            newCategoryHTML = `<span class="category category2">${ctx.categories[key].category}</span>`
        }
        colorCategory.innerHTML = newCategoryHTML

        let newWordsHTML = "tbd"
        const wordsLen = ctx.categories[key].words.length
        if (wordsLen > 0) {
            newWordsHTML = ""
            for (let i = 0; i < wordsLen; i++) {
                newWordsHTML += `<span class="category category2">${ctx.categories[key].words[i]}</span>`
            }
        }
        colorWords.innerHTML = newWordsHTML
    }
}

displayCategoryAndWords()

/**
 * When the edit svg is clicked, call editColor for the parents color.
 * Change the edit icon to an x if not selected, or back to edit icon if already selected.
 * @param {HTMLElement} el 
 */
function editCategories(el) {
    const color = el.parentElement.parentElement
    if (color.classList[1] == ctx.state) {
        el.src = "/static/edit.svg"
        editColor(color)
    } else if (ctx.state == States.NONE) {
        el.src = "/static/x.svg"
        editColor(color)
    }
}

/**
 * Update the context state from NONE to a specific color, give the color a border, and make the category/words inputs editable.
 * Updates the context state from a specific color to NONE, remove the colors border, and make the category/words inputs disabled.
 * @param {string} color 
 */
function editColor(color) {
    if (ctx.state === States.NONE) {
        ctx.state = color.classList[1]
        color.classList.add("color-selected")
        categoryInput.disabled = false
        wordsInput.disabled = false
        saveCategoriesButton.disabled = false
        populateCategoriesAndWords()
        categoryInput.focus()
    } else {
        ctx.state = States.NONE
        color.classList.remove("color-selected")
        categoryInput.disabled = true
        wordsInput.disabled = true
        saveCategoriesButton.disabled = true
        removeCategoriesAndWords()
    }
}

/**
 * Populates all of the saved categories and words for a specific color in the category input and words input,
 * also 
 */
function populateCategoriesAndWords() {
    // set category
    if (ctx.categories[ctx.state].category != "") {
        createCategory(ctx.categories[ctx.state].category, categoryInput)
    }

    // set words
    for (let i = 0; i < ctx.categories[ctx.state].words.length; i++) {
        createWord(ctx.categories[ctx.state].words[i], wordsInput)
    }
}

/**
 * Removes all of the populated categories and words from the category input and words input,
 * and remove them from currentCategory and currentWords.
 */
function removeCategoriesAndWords() {
    for (let i = 0; i < currentCategory.length; i++) {
        removeCategory(currentCategory[i].children[0])
    }
    for (let i = 0; i < currentWords.length; i++) {
        removeWord(currentWords[i].children[0], false)
    }
    currentCategory.splice(0, currentCategory.length)
    currentWords.splice(0, currentWords.length)
    categoryInput.value = ""
    wordsInput.value = ""
}

/**
 * 
 */
function checkSubmitStatus() {
    if (ctx.categories.yellow.words.length == 4
        && ctx.categories.green.words.length == 4
        && ctx.categories.blue.words.length == 4
        && ctx.categories.purple.words.length == 4) {
        submitButton.disabled = false
    } else {
        submitButton.disabled = true
    }
}

checkSubmitStatus()


function saveCurrentCategories() {
    if (ctx.state != States.NONE) {
        let categoryString = currentCategory.length > 0 ? currentCategory[0].innerText : ""
        ctx.categories[ctx.state].category = categoryString.slice(0, categoryString.indexOf("\n"))
        ctx.categories[ctx.state].words.splice(0, ctx.categories[ctx.state].words.length)
        for (let i = 0; i < currentWords.length; i++) {
            let wordString = currentWords[i].innerText
            ctx.categories[ctx.state].words.push(wordString.slice(0, wordString.indexOf("\n")))
        }

        //get gameId
        const gameId = document.getElementById("gameId-input")
        ctx.gameId = gameId != null ? gameId.value.trim().replace(/\s+/g, "-") : ""

        //clear category/word input
        removeCategoriesAndWords()

        //revert edit icon and set state to none
        const currentDiv = document.querySelector(`div.colors-item.${ctx.state}`)
        currentDiv.querySelector("img").src = "/static/edit.svg"
        editColor(currentDiv)

        // Save ctx
        localStorage.setItem("ctx", JSON.stringify(ctx))

        //display words + category
        displayCategoryAndWords()

        //enable submit button if ready
        checkSubmitStatus()
    }
}

/**
 * Hide the warning message div
 * @param {HTMLElement} el 
 */
function closeWarning(el) {
    warningDiv = el.parentElement
    warningDiv.style.visibility = "hidden"
}

/**
 * Returns the current categories present in context
 * @returns {Array<string>} category array
 */
function getCategoriesFromContext() {
    let categories = []
    if (ctx.categories.yellow.category) {
        categories.push(ctx.categories.yellow.category)
    }
    if (ctx.categories.green.category) {
        categories.push(ctx.categories.green.category)
    }
    if (ctx.categories.blue.category) {
        categories.push(ctx.categories.blue.category)
    }
    if (ctx.categories.purple.category) {
        categories.push(ctx.categories.purple.category)
    }

    return categories
}

/**
 * Attempt to submit the categories and words to create a game,
 * if unsuccessful, display an error message on screen with the failure reason
 * if successful, store the meta data in local storage and redirect to the game url
 * @returns {null}
 */
async function submitCategories() {
    const url = ""
    try {
        let body = ctx.categories
        body.gameId = ctx.gameId ?? ""
    
        const response = await fetch(url, {
            headers: {
                "Accept": "application/json",
                "Content-Type": "application/json",
            },
            method: "POST",
            body: JSON.stringify(body)
        })
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`)
        }

        const json = await response.json()
        if (json.success === false) {
            const msg = document.getElementById("warning-message")
            msg.innerText = json.failureReason
            msg.parentElement.style.visibility = "visible"
            submitButton.disabled = true
        } else if (json.gameId && json.gameId != "") {
            // localStorage.removeItem("ctx")
            storeGame(json.gameId, Date.now(), getCategoriesFromContext())
            window.location.href = `/game/${json.gameId}/`
            return
        }
        // console.log(json)
    } catch (error) {
        console.error(error.message);
    }
}

/**
 * Event listener for key presses when the context state is "NONE"
 * The following keys are listened for:
 * - Y/y -> set state to "yellow" and make the category and words inputs available
 * - G/g -> set state to "green" and make the category and words inputs available
 * - B/b -> set state to "blue" and make the category and words inputs available
 * - P/p -> set state to "purple" and make the category and words inputs available
 */
document.addEventListener("keydown", function (event) {
    if (ctx.state != States.NONE) {
        return
    }
    const gameIdInput = document.getElementById("gameId-input")
    if (gameIdInput == document.activeElement) {
        return
    }
    if (event.key === "Y" || event.key === "y") {
        if (ctx.state == States.NONE) {
            const yellow = document.querySelector("div.colors-item.yellow")
            const img = yellow.querySelector("img")
            // console.log(img)
            yellow.querySelector("img").click()
        }
        event.preventDefault()
    } else if (event.key === "G" || event.key === "g") {
        if (ctx.state == States.NONE) {
            const green = document.querySelector("div.colors-item.green")
            green.querySelector("img").click()
        }
        event.preventDefault()
    } else if (event.key === "B" || event.key === "b") {
        if (ctx.state == States.NONE) {
            const blue = document.querySelector("div.colors-item.blue")
            blue.querySelector("img").click()
        }
        event.preventDefault()
    } else if (event.key === "P" || event.key === "p") {
        if (ctx.state == States.NONE) {
            const purple = document.querySelector("div.colors-item.purple")
            purple.querySelector("img").click()
        }
        event.preventDefault()
    }
}, true)

const WORD_CHAR_LIMIT = 20

/**
 * @returns {bool} false if word-input value is greater than max allowed characters, true otherwise
 */
function isWordsInputValid() {
    const input = wordsInput.value
    if (input.length > WORD_CHAR_LIMIT) {
        return false
    }
    return true
}

/**
 * Event listener for key up event in the word input when the context state is "YELLOW", "GREEN", "BLUE", or "PURPLE"
 * If the length of the input value is greater than 20 chars, set the outline to red, otherwise keep as default blue.
 */
document.getElementById("word-input").addEventListener("keyup", function () {
    const input = this.value.trim();
    if (input.length > WORD_CHAR_LIMIT) {
        if (!wordsInput.classList.contains("toolong")) {
            wordsInput.classList.add("toolong")
        }
    } else {
        wordsInput.classList.remove("toolong")
    }
})

//abcdefghijklmnopqrstuv
/**
 * Event listener for key presses in the word input when the context state is "YELLOW", "GREEN", "BLUE", or "PURPLE"
 * The following keys are listened for:
 * - ,/enter -> if there is an input available, convert the input into a saved word
 * - backspace/delete -> if there is input text, remove the right most word
 */
document.getElementById("word-input").addEventListener("keydown", function (event) {
    const input = this.value.trim();

    if (event.key === "," || event.key === "Enter") {
        event.preventDefault();
        if (!isWordsInputValid()) {
            return
        }
        if (input && currentWords.length < 4) {
            createWord(input, this);
            this.value = "";
        }
    } else if (event.key === "Backspace" || event.key === "Delete") {
        if (!input) {
            const mostRecentWord = currentWords.pop()
            removeWord(mostRecentWord.children[0], false)
            event.preventDefault();
        }
    }
})

/**
 * Creates a word for a specific state.  Put the word in the currentWords array.
 * @param {string} text the text to save
 * @param {HTMLElement} inputElement the input element to insert a div before
 */
function createWord(text, inputElement) {
    const wordContainer = document.querySelector(".word-container");
    const word = document.createElement("div");
    word.className = "category";
    word.innerHTML = `${text}<span class="word-remove" onclick="removeWord(this)">&times;</span>`;
    wordContainer.insertBefore(word, inputElement);
    currentWords.push(word)
}

/**
 * Removes a word from a specific state
 * @param {HTMLElement} element the element to remove a word for
 * @param {bool} findWord true if the word needs to be searched for and removed from the currentWords array, false otherwise
 */
function removeWord(element, findWord = true) {
    if (findWord) {
        let idx = currentWords.indexOf(element.parentElement)
        currentWords.splice(idx, 1)
    }
    element.parentElement.remove();
    wordsInput.focus()
}

const CATEGORY_CHAR_LIMIT = 40

/**
 * @returns {bool} false if category-input value is greater than the max allowed characters, true otherwise
 */
function isCategoryInputValid() {
    const input = categoryInput.value
    if (input.length > CATEGORY_CHAR_LIMIT) {
        return false
    }
    return true
}

/**
 * Event listener for key up event in the word input when the context state is "YELLOW", "GREEN", "BLUE", or "PURPLE"
 * If the length of the input value is greater than 40 chars, set the outline to red, otherwise keep as default blue.
 */
document.getElementById("category-input").addEventListener("keyup", function () {
    if (!isCategoryInputValid()) {
        if (!categoryInput.classList.contains("toolong")) {
            categoryInput.classList.add("toolong")
        }
    } else {
        categoryInput.classList.remove("toolong")
    }
})

/**
 * Event listener for key presses in the category input when the context state is "YELLOW", "GREEN", "BLUE", or "PURPLE"
 * The following keys are listened for:
 * - ,/enter -> if there is an input available, convert the input into a saved category
 * - backspace/delete -> if there is input text, remove the right most category
 */
document.getElementById("category-input").addEventListener("keydown", function (event) {
    if (event.key === "," || event.key === "Enter") {
        event.preventDefault();
        if (!isCategoryInputValid()) {
            return
        }
        const input = this.value.trim();
        if (input && currentCategory.length < 1) {
            createCategory(input, this);
            this.value = "";
        }
    } else if (event.key === "Backspace" || event.key === "Delete") {
        const input = this.value.trim();
        if (!input) {
            const mostRecentCategory = currentCategory.pop()
            removeCategory(mostRecentCategory.children[0])
            event.preventDefault();
        }
    }
})

/**
 * Creates a category for a specific state, put the category in currentCategory array.
 * @param {string} text the text to save
 * @param {HTMLElement} inputElement the input element to insert a div before
 */
function createCategory(text, inputElement) {
    const categoryContainer = document.querySelector(".category-container");
    const category = document.createElement("div");
    category.className = "category";
    category.innerHTML = `${text}<span class="word-remove" onclick="removeCategory(this)">&times;</span>`;
    categoryContainer.insertBefore(category, inputElement);
    currentCategory.push(category)
}

/**
 * Removes a word from a specific state
 * @param {HTMLElement} element the element to remove a word for
 */
function removeCategory(element) {
    element.parentElement.remove();
    categoryInput.focus()
}