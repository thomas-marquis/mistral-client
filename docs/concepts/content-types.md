# Content Types

When we chat with an AI, with usually exchange messages (system, user, assistant or tool).
A message carries a content, which can be either:
- a simple string
- one or more chunks

Here are some examples of API responses in both formats:

=== "String content"

    ```json
    {
        "role": "assistant",
        "content": "Hello, I'm a useful assistant."
    }
    ```

=== "Chunked content"

    ```json
    {
        "role": "user",
        "content": [
            {"type": "text", "text": "please describe this image"},
            {"image_url": "https://example.com/image.png", "type": "image_url"}
        ]
    }
    ```

Chunks can be used to send complex data types to Mistral: images, documents, audio...
Each role can handle different types of chunks:

| Chunk type \ roles | User                      | Assistant                 | System                    | Tool                       |
|--------------------|---------------------------|---------------------------|---------------------------|----------------------------|
| Text               | :fontawesome-solid-check: | :fontawesome-solid-check: | :fontawesome-solid-check: | :fontawesome-solid-check:  |
| Image URL          | :fontawesome-solid-check: | :fontawesome-solid-check: |                           | :fontawesome-solid-check:  |
| Document URL       | :fontawesome-solid-check: | :fontawesome-solid-check: |                           | :fontawesome-solid-check:  |
| File               | :fontawesome-solid-check: | :fontawesome-solid-check: |                           | :fontawesome-solid-check:  |
| Reference          | :fontawesome-solid-check: | :fontawesome-solid-check: |                           | :fontawesome-solid-check:  |
| Thinking           | :fontawesome-solid-check: | :fontawesome-solid-check: | :fontawesome-solid-check: | :fontawesome-solid-check:  |
| Audio              | :fontawesome-solid-check: | :fontawesome-solid-check: |                           | :fontawesome-solid-check:  |

All roles can handle simple string content.

So, when you create a message, depending on your use case, you have to choose between a content string or an array of chunks. 