document.addEventListener("DOMContentLoaded", () => {
    const searchInput = document.querySelector(
        "#navbar-search-input",
    );

    const searchResults = document.querySelector(
        "#navbar-search-results",
    );

    const accountToggle = document.querySelector(
        "[data-account-menu-toggle]",
    );

    const accountDropdown = document.querySelector(
        "[data-account-menu-dropdown]",
    );

    let searchTimer = null;
    let requestController = null;

    const escapeHTML = (value) => {
        const element = document.createElement("div");

        element.textContent = value;

        return element.innerHTML;
    };

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

    const renderSuggestions = (
        query,
        categories,
        posts,
    ) => {
        if (!searchResults) {
            return;
        }

        let html = "";

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

    const search = async (query) => {
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

            if (!response.ok) {
                closeSearch();
                return;
            }

            const data = await response.json();

            renderSuggestions(
                query,
                data.categories || [],
                data.posts || [],
            );
        } catch (error) {
            if (error.name !== "AbortError") {
                closeSearch();
            }
        }
    };

    if (searchInput && searchResults) {
        searchInput.addEventListener(
            "input",
            () => {
                const query =
                    searchInput.value.trim();

                clearTimeout(searchTimer);

                if (query.length < 2) {
                    closeSearch();
                    return;
                }

                searchTimer = setTimeout(
                    () => {
                        search(query);
                    },
                    180,
                );
            },
        );

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

                closeSearch();
            },
        );
    }

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