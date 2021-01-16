export class App {
    constructor(root: HTMLElement) {
        const h1 = document.createElement("h1")
        h1.textContent = "Hello World"
        root.appendChild(h1)
    }
}