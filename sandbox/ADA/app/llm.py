"""
This module interacts with the Azure OpenAI service to send and receive messages.
Functions include:
    get_system_message_from_file(filename: str) -> dict:
        Reads a system prompt from a file and returns it in the required message format.
    get_system_message_from_string(system_prompt: str) -> dict:
        Converts a system prompt string into the required message format.
    send_message(messages: list, is_streaming: bool = True) -> Union[str, Any]:
        Sends a list of messages to the Azure OpenAI service and returns the response.
        Handles errors and retries if necessary.
Usage:
    This script can be run directly to test the LLM interaction. It sends a predefined system message and user message to the Azure OpenAI service and prints the response.
"""

import os
from openai import AzureOpenAI, BadRequestError

endpoint = os.getenv("AZURE_OPENAI_ENDPOINT")
model_name = os.getenv("AZURE_OPENAI_MODEL_NAME")
deployment = os.getenv("AZURE_OPENAI_DEPLOYMENT")
subscription_key = os.getenv("AZURE_OPENAI_API_KEY")
api_version = "2024-12-01-preview"

client = AzureOpenAI(
    api_version=api_version,
    azure_endpoint=endpoint,
    api_key=subscription_key,
)

def get_system_message_from_file(filename):
    system_prompt = ""
    with open(filename, "r") as file:
        system_prompt = file.read()

    return get_system_message_from_string(system_prompt)

def get_system_message_from_string(system_prompt):
    system_status = None
    system_status = {
            "role": "system", 
            "content": system_prompt
        }
    return system_status


def send_message(messages, is_streaming=True):
    """
    Sends a message to the chat completion client and handles potential errors.
    Args:
        messages (list): A list of message dictionaries, each containing a 'content' key with a string value.
        is_streaming (bool, optional): Flag to indicate if the response should be streamed. Defaults to True.
    Raises:
        ValueError: If any message content is not a string.
    Returns:
        response: The response from the chat completion client if successful.
        str: An error message if an exception occurs.
    """
    # Ensure all messages have the correct format
    for message in messages:
        if not isinstance(message['content'], str):
            raise ValueError(f"Invalid content type: {type(message['content'])}. Expected a string.")
    
    try:
        response = client.chat.completions.create(
                        stream=True,
                        messages=messages,
                        max_tokens=4096,
                        temperature=1.0,
                        top_p=1.0,
                        model=deployment,
                    )
    except Exception as e:
        if isinstance(e, BadRequestError):
            error_code = e.error.get('code')
            if error_code == 'content_filter':
                print(f" triggered with prompt {messages[-1]['content']}. Modifying prompt and retrying...")
            else:
                print(f"Error: {e.error['message']}\n\nCode: {error_code}\n\nprompt: {messages[-1]['content']}")
        else:
            print(f"An unexpected error occurred: {str(e)} with prompt {messages[-1]['content']}")
        messages.pop()
        return "I'm sorry, but I'm not sure how to respond to that, try rephrasing."
    
    return response

if __name__ == "__main__":
    print ("Testing LLM...")

    messages = []
    messages.append(get_system_message_from_string("You re a grumpy agent in the style of Marvin the Paranoid Android."))
    messages.append({
        "role": "user", 
        "content": "Say Hello."
    })
    
    print ("Sending message...")
    response = send_message(messages, False)

    print("Response:")
    if isinstance(response, str):
        print(response)
    else:
        for chunk in response:
            if chunk.choices:
                if chunk.choices[0].delta.content:
                    print(chunk.choices[0].delta.content, end='', flush=True)
    print()