document.addEventListener("DOMContentLoaded", () => {
    const filters = document.querySelectorAll(".discussion-filter");
    const discussionSection = document.querySelector("#discussions");
    const discussionList = document.querySelector(".discussion-list");
    const discussionTitle = document.querySelector(".discussion-title");

    if (
        filters.length === 0 ||
        !discussionSection ||
        !discussionList ||
        !discussionTitle
    ) {
        return;
    }

    filters.forEach((filter) => {
        filter.addEventListener("click", async (event) => {
            event.preventDefault();

            const url = filter.href;

            try {
                filters.forEach((item) => {
                    item.classList.add("is-loading");
                });

                const response = await fetch(url, {
                    headers: {
                        "X-Requested-With": "fetch",
                    },
                });

                if (!response.ok) {
                    throw new Error("Failed to load posts");
                }

                const html = await response.text();

                const parser = new DOMParser();
                const documentResult = parser.parseFromString(
                    html,
                    "text/html"
                );

                const newDiscussionList =
                    documentResult.querySelector(".discussion-list");

                const newDiscussionTitle =
                    documentResult.querySelector(".discussion-title");

                if (!newDiscussionList || !newDiscussionTitle) {
                    throw new Error("Invalid filter response");
                }

                discussionList.innerHTML =
                    newDiscussionList.innerHTML;

                discussionTitle.textContent =
                    newDiscussionTitle.textContent;

                filters.forEach((item) => {
                    item.classList.remove("is-active");
                });

                filter.classList.add("is-active");

                window.history.replaceState(
                    {},
                    "",
                    filter.getAttribute("href")
                );
            } catch (error) {
                console.error(error);
            } finally {
                filters.forEach((item) => {
                    item.classList.remove("is-loading");
                });
            }
        });
    });
});