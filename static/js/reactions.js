// Wait until the initial HTML document has been completely loaded and parsed.
document.addEventListener("DOMContentLoaded", () => {
    // Select all forms used for post and comment reactions.
    const reactionForms = document.querySelectorAll(".reaction-form");

    // Attach asynchronous submission behavior to every reaction form.
    reactionForms.forEach((form) => {
        form.addEventListener("submit", async (event) => {
            // Prevent the form from performing a normal page submission.
            event.preventDefault();

            const button = form.querySelector(".reaction-button");

            // Stop if the reaction button is missing or already disabled.
            if (!button || button.disabled) {
                return;
            }

            // Prevent repeated submissions while the request is in progress.
            button.disabled = true;

            try {
                // Send the reaction to the server without reloading the page.
                const response = await fetch(form.action, {
                    method: "POST",
                    body: new FormData(form),
                    headers: {
                        "X-Requested-With": "fetch",
                    },
                });

                // Treat unsuccessful HTTP responses as errors.
                if (!response.ok) {
                    throw new Error("Reaction request failed");
                }

                // Parse the updated reaction state and counts returned by the server.
                const result = await response.json();

                // Find the reaction controls that belong to the submitted form.
                const reactionContainer = form.closest(
                    ".reaction-group, .comment-reactions"
                );

                if (!reactionContainer) {
                    return;
                }

                // Find the like and dislike buttons inside the reaction group.
                const likeButton = reactionContainer.querySelector(
                    ".reaction-like"
                );

                const dislikeButton = reactionContainer.querySelector(
                    ".reaction-dislike"
                );

                if (!likeButton || !dislikeButton) {
                    return;
                }

                // Update the active state of the like button.
                likeButton.classList.toggle(
                    "is-active",
                    result.reaction === 1
                );

                // Update the active state of the dislike button.
                dislikeButton.classList.toggle(
                    "is-active",
                    result.reaction === -1
                );

                // Find the elements that display the current reaction counts.
                const likeCount = likeButton.querySelector(
                    ".reaction-count"
                );

                const dislikeCount = dislikeButton.querySelector(
                    ".reaction-count"
                );

                // Update the displayed number of likes.
                if (likeCount) {
                    likeCount.textContent = result.likes;
                }

                // Update the displayed number of dislikes.
                if (dislikeCount) {
                    dislikeCount.textContent = result.dislikes;
                }
            } catch (error) {
                // Log request or response errors without breaking the page.
                console.error(error);
            } finally {
                // Re-enable the reaction button when the request finishes.
                button.disabled = false;
            }
        });
    });
});