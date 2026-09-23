// Wait until the initial HTML document has been completely loaded and parsed.
document.addEventListener("DOMContentLoaded", () => {
    // Select the navbar search input and its suggestions container.
    const searchInput = document.querySelector(
        "#navbar-search-input",
    );

    const searchResults = document.querySelector(
        "#navbar-search-results",
    );

    // Select the account menu toggle and dropdown.
    const accountToggle = document.querySelector(
        "[data-account-menu-toggle]",
    );

    const accountDropdown = document.querySelector(
        "[data-account-menu-dropdown]",
    );

    // Keep track of the search debounce timer and active request.
    let searchTimer = null;
    let requestController = null;

    // Escape text before inserting it into generated HTML.
    const escapeHTML = (value) => {
        const element = document.createElement("div");

        element.textContent = value;

        return element.innerHTML;
    };

    // Hide the search suggestions and update its accessibility state.
    const closeSearch = () => {
        if (!searchResults || !searchInput) {
            return;
        }

        searchResults.hidden = true;

        searchInput.setAttribute(
            "aria-expanded",
            "false",
        );
    };

    // Show the search suggestions and update its accessibility state.
    const openSearch = () => {
        if (!searchResults || !searchInput) {
            return;
        }

        searchResults.hidden = false;

        searchInput.setAttribute(
            "aria-expanded",
            "true",
        );
    };

    // Render category and discussion suggestions returned by the server.
    const renderSuggestions = (
        query,
        categories,
        posts,
    ) => {
        if (!searchResults) {
            return;
        }

        let html = "";

        // Add up to three matching category suggestions.
        if (categories.length > 0) {
            html += `
                <div class="search-suggestion-heading">
                    Categories
                </div>
            `;

            categories.slice(0, 3).forEach((category) => {
                html += `
                    <a
                        class="search-suggestion"
                        href="/categories/${encodeURIComponent(category.slug)}"
                        role="option"
                    >
                        <span class="search-suggestion-icon">
                            #
                        </span>

                        <span class="search-suggestion-content">
                            <span class="search-suggestion-title">
                                ${escapeHTML(category.name)}
                            </span>

                            <span class="search-suggestion-description">
                                ${escapeHTML(
                                    category.tagline ||
                                    category.description ||
                                    "Gaming category",
                                )}
                            </span>
                        </span>
                    </a>
                `;
            });
        }

        // Add up to five matching discussion suggestions.
        if (posts.length > 0) {
            html += `
                <div class="search-suggestion-heading">
                    Discussions
                </div>
            `;

            posts.slice(0, 5).forEach((post) => {
                html += `
                    <a
                        class="search-suggestion"
                        href="/posts/${post.id}"
                        role="option"
                    >
                        <span class="search-suggestion-icon">
                            ◇
                        </span>

                        <span class="search-suggestion-content">
                            <span class="search-suggestion-title">
                                ${escapeHTML(post.title)}
                            </span>

                            <span class="search-suggestion-description">
                                by ${escapeHTML(post.username)}
                            </span>
                        </span>
                    </a>
                `;
            });
        }

        // Display an empty-state message when nothing matches the query.
        if (
            categories.length === 0 &&
            posts.length === 0
        ) {
            html = `
                <div class="search-suggestion-empty">
                    No results found for
                    "${escapeHTML(query)}"
                </div>
            `;
        }

        // Always provide a link to the complete search results page.
        html += `
            <a
                class="search-suggestion-all"
                href="/search?q=${encodeURIComponent(query)}"
            >
                View all results for "${escapeHTML(query)}"
            </a>
        `;

        searchResults.innerHTML = html;

        openSearch();
    };

    // Fetch search suggestions for the current query.
    const search = async (query) => {
        // Cancel the previous request when a newer search starts.
        if (requestController) {
            requestController.abort();
        }

        requestController =
            new AbortController();

        try {
            const response = await fetch(
                `/search/suggestions?q=${encodeURIComponent(query)}`,
                {
                    signal:
                        requestController.signal,
                    headers: {
                        Accept: "application/json",
                    },
                },
            );

            // Hide the suggestions if the server returns an unsuccessful response.
            if (!response.ok) {
                closeSearch();
                return;
            }

            const data = await response.json();

            // Render the suggestions returned by the server.
            renderSuggestions(
                query,
                data.categories || [],
                data.posts || [],
            );
        } catch (error) {
            // Ignore intentionally aborted requests and close the search on other errors.
            if (error.name !== "AbortError") {
                closeSearch();
            }
        }
    };

    // Enable live search when the required search elements are available.
    if (searchInput && searchResults) {
        searchInput.addEventListener(
            "input",
            () => {
                const query =
                    searchInput.value.trim();

                // Reset the previous debounce timer whenever the input changes.
                clearTimeout(searchTimer);

                // Do not search until at least two characters have been entered.
                if (query.length < 2) {
                    closeSearch();
                    return;
                }

                // Delay the request slightly to avoid searching after every keystroke.
                searchTimer = setTimeout(
                    () => {
                        search(query);
                    },
                    180,
                );
            },
        );

        // Reopen existing suggestions when the search input receives focus.
        searchInput.addEventListener(
            "focus",
            () => {
                if (
                    searchInput.value.trim().length >= 2 &&
                    searchResults.innerHTML.trim() !== ""
                ) {
                    openSearch();
                }
            },
        );
    }

    // Toggle the account dropdown when its button is clicked.
    if (accountToggle && accountDropdown) {
        accountToggle.addEventListener(
            "click",
            () => {
                const isOpen =
                    accountToggle.getAttribute(
                        "aria-expanded",
                    ) === "true";

                accountToggle.setAttribute(
                    "aria-expanded",
                    String(!isOpen),
                );

                accountDropdown.hidden = isOpen;

                // Close search suggestions while the account menu is being used.
                closeSearch();
            },
        );
    }

    // Close open navbar controls when the user clicks outside them.
    document.addEventListener(
        "click",
        (event) => {
            const searchWrapper =
                event.target.closest(
                    ".navbar-search-wrapper",
                );

            if (!searchWrapper) {
                closeSearch();
            }

            if (
                accountToggle &&
                accountDropdown
            ) {
                const accountMenu =
                    event.target.closest(
                        ".account-menu",
                    );

                if (!accountMenu) {
                    accountDropdown.hidden = true;

                    accountToggle.setAttribute(
                        "aria-expanded",
                        "false",
                    );
                }
            }
        },
    );

    // Close search suggestions and the account menu when Escape is pressed.
    document.addEventListener(
        "keydown",
        (event) => {
            if (event.key !== "Escape") {
                return;
            }

            closeSearch();

            if (
                accountToggle &&
                accountDropdown
            ) {
                accountDropdown.hidden = true;

                accountToggle.setAttribute(
                    "aria-expanded",
                    "false",
                );
            }
        },
    );
});