import { isErrorWithProperty } from "./common.js"

/**
 * Contains the possible states for creating a connections game.
 */
enum States {
    NONE = "none",
    YELLOW = "yellow",
    GREEN = "green",
    BLUE = "blue",
    PURPLE = "purple",
}

/**
 * Gets the State enum key from value string
 */
const statesReverseMap: { [value: string]: States } = {
    "none": States.NONE,
    "yellow": States.YELLOW,
    "green": States.GREEN,
    "blue": States.BLUE,
    "purple": States.PURPLE
}

/**
 * Represents a single category with the words associated to that category.
 */
type ColorCategory = {
    category: string
    words: Array<string>
}

/**
 * Represents the four connections categories and words related to each category.
 */
type Categories = {
    yellow: ColorCategory
    green: ColorCategory
    blue: ColorCategory
    purple: ColorCategory
}

type CreateBody<Categories> = Categories & { [key: string]: any };

/**
 * Represents the context of the page storing the state, categories, and custom game id.
 */
type Context = {
    state: States
    categories: Categories
    gameId: string
}

/**
 * Creates an empty object for storing the state of each connection color.
 * @returns {Categories}
 */
function getEmptyCategories(): Categories {
    return {
        yellow: {
            category: "",
            words: []
        },
        green: {
            category: "",
            words: []
        },
        blue: {
            category: "",
            words: []
        },
        purple: {
            category: "",
            words: []
        }
    }
}

/**
 * Creates a default context object
 * @returns {Context}
 */
function getDefaultContext(): Context {
    return {
        state: States.NONE,
        categories: getEmptyCategories(),
        gameId: ""
    }
}

/**
 * Stores meta data for the users created games in local storage
 * @param {string} gameId 
 * @param {number} timestamp 
 * @param {Array<string>} categories 
 */
function storeGame(gameId: string, timestamp: number, categories: Array<string>): void {
    const game = {
        "gameId": gameId,
        "timestamp": timestamp,
        "categories": categories,
    }
    createdGames.push(game)
    localStorage.setItem("createdGames", JSON.stringify(createdGames))
}

const yellowEdit = document.getElementById("yellow-edit") ?? (() => { throw new Error("yellowEdit cannot be null") })()
yellowEdit.addEventListener("click", (el) => editCategories(el.target as HTMLImageElement))

const greenEdit = document.getElementById("green-edit") ?? (() => { throw new Error("greenEdit cannot be null") })()
greenEdit.addEventListener("click", (el) => editCategories(el.target as HTMLImageElement))

const blueEdit = document.getElementById("blue-edit") ?? (() => { throw new Error("blueEdit cannot be null") })()
blueEdit.addEventListener("click", (el) => editCategories(el.target as HTMLImageElement))

const purpleEdit = document.getElementById("purple-edit") ?? (() => { throw new Error("purpleEdit cannot be null") })()
purpleEdit.addEventListener("click", (el) => editCategories(el.target as HTMLImageElement))


/**
 * When the edit svg is clicked, call editColor for the parents color.
 * Change the edit icon to an x if not selected, or back to edit icon if already selected.
 * @param {HTMLImageElement} el 
 */
function editCategories(el: HTMLImageElement) {
    const color = el?.parentElement?.parentElement
    if (color && color.classList[1] == ctx.state) {
        el.src = "/static/edit.svg"
        editColor(color)
    } else if (color && ctx.state == States.NONE) {
        el.src = "/static/x.svg"
        editColor(color)
    }
}


var categoryInput: HTMLInputElement
var wordsInput: HTMLInputElement
var saveCategoriesButton: HTMLButtonElement
var submitButton: HTMLButtonElement
var createdGames: Array<any>
var ctx: Context

var currentWords: Array<HTMLDivElement> = []
var currentCategory: Array<HTMLDivElement> = []

/**
 * Setup the create script
 */
function setupCreate() {
    categoryInput = document.getElementById("category-input") as HTMLInputElement ?? (() => { throw new Error("categoryInput cannot be null") })()
    wordsInput = document.getElementById("word-input") as HTMLInputElement ?? (() => { throw new Error("wordsInput cannot be null") })()
    saveCategoriesButton = document.getElementById("save-categories-button") as HTMLButtonElement ?? (() => { throw new Error("saveCategoriesButton cannot be null") })()
    submitButton = document.getElementById("submit") as HTMLButtonElement ?? (() => { throw new Error("submitButton cannot be null") })()

    let ctxStorage = localStorage.getItem("ctx")

    if (!ctxStorage) {
        ctx = getDefaultContext()
        localStorage.setItem("ctx", JSON.stringify(ctx))
    } else {
        ctx = JSON.parse(ctxStorage)
        if (ctx.state !== States.NONE) {
            ctx.state = States.NONE
            localStorage.setItem("ctx", JSON.stringify(ctx))
        }
    }

    createdGames = JSON.parse(localStorage.getItem("createdGames") ?? "[]")
    if (!createdGames) {
        localStorage.setItem("createdGames", JSON.stringify(createdGames))
    }

    displayCategoryAndWords()

    /**
     * Event listener for key up event in the word input when the context state is "YELLOW", "GREEN", "BLUE", or "PURPLE"
     * If the length of the input value is greater than 40 chars, set the outline to red, otherwise keep as default blue.
     */
    document.getElementById("category-input")?.addEventListener("keyup", function () {
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
    document.getElementById("category-input")?.addEventListener("keydown", function (event) {
        const that = this as HTMLInputElement
        if (event.key === "," || event.key === "Enter") {
            event.preventDefault();
            if (!isCategoryInputValid()) {
                return
            }
            const input = that.value.trim();
            if (input && currentCategory.length < 1) {
                createCategory(input, that, true);
                that.value = "";
            }
        } else if (event.key === "Backspace" || event.key === "Delete") {
            const input = that.value.trim();
            if (!input) {
                const mostRecentCategory = currentCategory.pop()
                removeCategory(mostRecentCategory?.children[0] as HTMLElement, true)
                event.preventDefault();
            }
        }
    })

    /**
     * Event listener for key up event in the word input when the context state is "YELLOW", "GREEN", "BLUE", or "PURPLE"
     * If the length of the input value is greater than 20 chars, set the outline to red, otherwise keep as default blue.
     */
    document.getElementById("word-input")?.addEventListener("keyup", function () {
        const that = this as HTMLInputElement

        const input = that.value.trim();
        if (input.length > WORD_CHAR_LIMIT) {
            if (!wordsInput.classList.contains("toolong")) {
                wordsInput.classList.add("toolong")
            }
        } else {
            wordsInput.classList.remove("toolong")
        }
    })

    /**
     * Event listener for key presses in the word input when the context state is "YELLOW", "GREEN", "BLUE", or "PURPLE"
     * The following keys are listened for:
     * - ,/enter -> if there is an input available, convert the input into a saved word
     * - backspace/delete -> if there is input text, remove the right most word
     */
    document.getElementById("word-input")?.addEventListener("keydown", function (event: KeyboardEvent) {
        const that: HTMLInputElement = this as HTMLInputElement
        const input: string = that.value.trim();

        if (event.key === "," || event.key === "Enter") {
            event.preventDefault();
            if (!isWordsInputValid()) {
                return
            }
            if (input && currentWords.length < 4) {
                createWord(input, that, true);
                that.value = "";
            }
        } else if (event.key === "Backspace" || event.key === "Delete") {
            if (!input) {
                const mostRecentWord = currentWords.pop()
                if (mostRecentWord && mostRecentWord.hasChildNodes()) {
                    removeWord(mostRecentWord.children[0] as HTMLElement, false, true)
                    event.preventDefault();
                }
            }
        }
    })

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
                yellow?.querySelector("img")?.click()
            }
            event.preventDefault()
        } else if (event.key === "G" || event.key === "g") {
            if (ctx.state == States.NONE) {
                const green = document.querySelector("div.colors-item.green")
                green?.querySelector("img")?.click()
            }
            event.preventDefault()
        } else if (event.key === "B" || event.key === "b") {
            if (ctx.state == States.NONE) {
                const blue = document.querySelector("div.colors-item.blue")
                blue?.querySelector("img")?.click()
            }
            event.preventDefault()
        } else if (event.key === "P" || event.key === "p") {
            if (ctx.state == States.NONE) {
                const purple = document.querySelector("div.colors-item.purple")
                purple?.querySelector("img")?.click()
            }
            event.preventDefault()
        }
    }, true)

    checkSubmitStatus()
}

setupCreate()

/**
 * Displays all of the categories and words for each color group from context.
 */
function displayCategoryAndWords() {
    for (const key in ctx["categories"]) {
        const colorCategory = document.getElementById(`category-${key}`)
        const colorWords = document.getElementById(`words-${key}`)
        let newCategoryHTML = "tbd"

        if (ctx["categories"].hasOwnProperty(key)) {
            const color = key as keyof Categories
            if (ctx["categories"][color].category != "") {
                newCategoryHTML = `<span class="category category2">${ctx["categories"][color].category}</span>`
            }
            if (colorCategory) {
                colorCategory.innerHTML = newCategoryHTML
            }
            let newWordsHTML = "tbd"
            const wordsLen = ctx["categories"][color].words.length
            if (wordsLen > 0) {
                newWordsHTML = ""
                for (let i = 0; i < wordsLen; i++) {
                    newWordsHTML += `<span class="category category2">${ctx["categories"][color].words[i]}</span>`
                }
            }
            if (colorWords) {
                colorWords.innerHTML = newWordsHTML
            }
        }
    }
}

/**
 * Update the context state from NONE to a specific color, give the color a border, and make the category/words inputs editable.
 * Updates the context state from a specific color to NONE, remove the colors border, and make the category/words inputs disabled.
 * @param {HTMLElement} color 
 */
function editColor(color: HTMLElement) {
    if (ctx.state === States.NONE) {
        let colorVal: string = color.classList[1]

        ctx.state = statesReverseMap[colorVal]
        color.classList.add("color-selected")
        if (categoryInput) {
            categoryInput.disabled = false
        }
        wordsInput.disabled = false
        saveCategoriesButton.disabled = false
        populateCategoriesAndWords()
        if (currentCategory.length > 0) {
            wordsInput.focus()
        } else {
            categoryInput.focus()
        }
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
    const state = ctx.state.toString()

    if (state === "yellow" || state === "green" || state === "blue" || state === "purple") {
        const color: Categories[typeof state] = ctx["categories"][state]

        // set category
        if (color.category != "") {
            createCategory(color.category, categoryInput)
        }

        // set words
        for (let i = 0; i < color.words.length; i++) {
            createWord(color.words[i], wordsInput)
        }
    }
}

/**
 * Removes all of the populated categories and words from the category input and words input,
 * and remove them from currentCategory and currentWords.
 */
function removeCategoriesAndWords() {
    for (let i = 0; i < currentCategory.length; i++) {
        removeCategory(currentCategory[i].children[0] as HTMLElement)
    }
    for (let i = 0; i < currentWords.length; i++) {
        removeWord(currentWords[i].children[0] as HTMLElement, false)
    }
    currentCategory.splice(0, currentCategory.length)
    currentWords.splice(0, currentWords.length)
    categoryInput.value = ""
    wordsInput.value = ""
}

/**
 * Checks if submit button should be disabled.
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

const saveCategories = document.getElementById("save-categories-button") ?? (() => { throw new Error("saveCategories cannot be null") })()
saveCategories.addEventListener("click", () => saveCurrentCategories())

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
        const gameId = document.getElementById("gameId-input") as HTMLInputElement
        ctx.gameId = gameId != null ? gameId.value.trim().replace(/\s+/g, "-") : ""

        //clear category/word input
        removeCategoriesAndWords()

        //revert edit icon and set state to none
        const currentDiv = document.querySelector(`div.colors-item.${ctx.state}`) as HTMLElement
        if (currentDiv) {
            const currentImg: HTMLImageElement = currentDiv.querySelector("img") as HTMLImageElement
            currentImg.src = "/static/edit.svg"
            editColor(currentDiv)
        }

        // Save ctx
        localStorage.setItem("ctx", JSON.stringify(ctx))

        //display words + category
        displayCategoryAndWords()

        //enable submit button if ready
        checkSubmitStatus()
    }
}

function saveCurrentCategoriesSilently(state: States) {
    if (state === States.NONE) {
        return
    }
    const color: Categories[typeof state] = ctx["categories"][state]

    let categoryString = currentCategory.length > 0 ? currentCategory[0].innerText : ""
    color.category = categoryString.slice(0, categoryString.indexOf("\n"))
    color.words.splice(0, color.words.length)
    for (let i = 0; i < currentWords.length; i++) {
        let wordString = currentWords[i].innerText
        color.words.push(wordString.slice(0, wordString.indexOf("\n")))
    }

    // Save ctx
    localStorage.setItem("ctx", JSON.stringify(ctx))

    //display words + category
    displayCategoryAndWords()
}

const submitWarningClose = document.getElementById("submit-warning-close") ?? (() => { throw new Error("submitWarningClose cannot be null") })()
submitWarningClose.addEventListener("click", (event) => closeWarning(event.target as HTMLElement))

/**
 * Hide the warning message div
 * @param {HTMLElement} el 
 */
function closeWarning(el: HTMLElement) {
    const warningDiv = el.parentElement
    if (warningDiv) {
        warningDiv.style.visibility = "hidden"
    }
}

/**
 * Returns the current categories present in context
 * @returns {Array<string>} category array
 */
function getCategoriesFromContext() {
    let categories: Array<string> = []
    const contextCategories: Categories = ctx["categories"]
    const yellow: ColorCategory = contextCategories["yellow"]
    const green: ColorCategory = contextCategories["green"]
    const blue: ColorCategory = contextCategories["blue"]
    const purple: ColorCategory = contextCategories["purple"]

    if (yellow["category"]) {
        categories.push(yellow["category"])
    }
    if (green["category"]) {
        categories.push(green["category"])
    }
    if (blue["category"]) {
        categories.push(blue["category"])
    }
    if (purple["category"]) {
        categories.push(purple["category"])
    }

    return categories
}

const submit = document.getElementById("submit") ?? (() => { throw new Error("submit cannot be null") })()
submit.addEventListener("click", () => submitCategories())

/**
 * Attempt to submit the categories and words to create a game,
 * if unsuccessful, display an error message on screen with the failure reason
 * if successful, store the meta data in local storage and redirect to the game url
 * @returns {null}
 */
async function submitCategories() {
    const url = ""
    try {
        const body: CreateBody<Categories> = {
            yellow: ctx["categories"].yellow,
            green: ctx["categories"].green,
            blue: ctx["categories"].blue,
            purple: ctx["categories"].purple,
            gameId: ctx.gameId ?? ""
        }

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
            if (msg) {
                msg.innerText = json.failureReason
                if (msg.parentElement) {
                    msg.parentElement.style.visibility = "visible"
                }
            }
            submitButton.disabled = true
        } else if (json.gameId && json.gameId != "") {
            localStorage.removeItem("ctx")
            storeGame(json.gameId, Date.now(), getCategoriesFromContext())
            window.location.href = `/game/${json.gameId}/`
            return
        }
    } catch (error: unknown) {
        if (isErrorWithProperty(error, "message")) {
            console.error(error.message)
        }
    }
}

const WORD_CHAR_LIMIT: number = 20

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
 * Creates a word for a specific state.  Put the word in the currentWords array.
 * @param {string} text the text to save
 * @param {HTMLInputElement} inputElement the input element to insert a div before
 * @param {boolean} shouldSave if we should silently save after creating word
 */
function createWord(text: string, inputElement: HTMLInputElement, shouldSave: boolean = false) {
    const wordContainer = document.querySelector(".word-container");
    const word = document.createElement("div");

    word.className = "category";

    const remove = document.createElement("span")
    remove.classList.add("word-remove")
    remove.innerHTML = "&times;"
    remove.addEventListener("click", (event) => removeWord(event.target as HTMLElement))
    word.innerHTML = text
    word.appendChild(remove)
    // word.innerHTML = `${text}`;
    if (wordContainer) {
        wordContainer.insertBefore(word, inputElement);
    }
    currentWords.push(word)
    if (shouldSave) {
        saveCurrentCategoriesSilently(ctx.state)
    }
}

/**
 * Removes a word from a specific state
 * @param {HTMLElement} element the element to remove a word for
 * @param {bool} findWord true if the word needs to be searched for and removed from the currentWords array, false otherwise
 * @param {boolean} shouldSave if we should silently save after removing word
 */
function removeWord(element: HTMLElement, findWord: boolean = true, shouldSave: boolean = false) {
    if (findWord) {
        let idx = currentWords.indexOf(element.parentElement as HTMLDivElement)
        currentWords.splice(idx, 1)
    }
    element?.parentElement?.remove();
    wordsInput.focus()
    if (shouldSave) {
        saveCurrentCategoriesSilently(ctx.state)
    }
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
 * Creates a category for a specific state, put the category in currentCategory array.
 * @param {string} text the text to save
 * @param {HTMLInputElement} inputElement the input element to insert a div before
 * @param {boolean} shouldSave if we should silently save after creating category
 */
function createCategory(text: string, inputElement: HTMLInputElement, shouldSave: boolean = false) {
    const categoryContainer = document.querySelector(".category-container") ?? (() => { throw new Error("category-container cannot be null") })()
    const category = document.createElement("div")
    category.className = "category"
    category.innerHTML = `${text}<span class="word-remove" onclick="removeCategory(this)">&times;</span>`
    categoryContainer.insertBefore(category, inputElement)
    currentCategory.push(category)
    if (shouldSave) {
        saveCurrentCategoriesSilently(ctx.state)
    }
}

/**
 * Removes a word from a specific state
 * @param {HTMLElement} element the element to remove a word for
 * @param {boolean} shouldSave if we should silently save after removing category
 */
function removeCategory(element: HTMLElement, shouldSave: boolean = false) {
    element.parentElement?.remove();
    categoryInput.focus()
    if (shouldSave) {
        saveCurrentCategoriesSilently(ctx.state)
    }
}