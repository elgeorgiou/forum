// Wait until the initial HTML document has been completely loaded and parsed.
document.addEventListener("DOMContentLoaded", () => {
    // Select all links used to filter the discussion list.
    const filters = document.querySelectorAll(".discussion-filter");

    // Select the discussion section and the elements that will be updated.
    const discussionSection = document.querySelector("#discussions");
    const discussionList = document.querySelector(".discussion-list");
    const discussionTitle = document.querySelector(".discussion-title");

    // Stop if the required discussion elements are not present on the page.
    if (
        filters.length === 0 ||
        !discussionSection ||
        !discussionList ||
        !discussionTitle
    ) {
        return;
    }

    // Attach the filtering behavior to every discussion filter.
    filters.forEach((filter) => {
        filter.addEventListener("click", async (event) => {
            // Prevent the browser from navigating to the link normally.
            event.preventDefault();

            const url = filter.href;

            try {
                // Mark all filters as loading while the new content is fetched.
                filters.forEach((item) => {
                    item.classList.add("is-loading");
                });

                // Request the filtered page without performing a full page reload.
                const response = await fetch(url, {
                    headers: {
                        "X-Requested-With": "fetch",
                    },
                });

                // Treat unsuccessful HTTP responses as errors.
                if (!response.ok) {
                    throw new Error("Failed to load posts");
                }

                // Read the returned HTML response as text.
                const html = await response.text();

                // Parse the returned HTML into a temporary document.
                const parser = new DOMParser();
                const documentResult = parser.parseFromString(
                    html,
                    "text/html"
                );

                // Extract the updated discussion list from the returned document.
                const newDiscussionList =
                    documentResult.querySelector(".discussion-list");

                // Extract the updated discussion title from the returned document.
                const newDiscussionTitle =
                    documentResult.querySelector(".discussion-title");

                // Stop if the response does not contain the expected elements.
                if (!newDiscussionList || !newDiscussionTitle) {
                    throw new Error("Invalid filter response");
                }

                // Replace the current discussion list with the filtered results.
                discussionList.innerHTML =
                    newDiscussionList.innerHTML;

                // Update the discussion title to match the selected filter.
                discussionTitle.textContent =
                    newDiscussionTitle.textContent;

                // Remove the active state from every filter.
                filters.forEach((item) => {
                    item.classList.remove("is-active");
                });

                // Mark the selected filter as active.
                filter.classList.add("is-active");

                // Update the browser URL without reloading the page.
                window.history.replaceState(
                    {},
                    "",
                    filter.getAttribute("href")
                );
            } catch (error) {
                // Log filtering errors without breaking the rest of the page.
                console.error(error);
            } finally {
                // Always remove the loading state when the request finishes.
                filters.forEach((item) => {
                    item.classList.remove("is-loading");
                });
            }
        });
    });
});