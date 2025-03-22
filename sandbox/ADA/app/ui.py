import os
import sys
import time
import traceback

import settings

info_line_count = 0

def get_user_input(prompt = ""):
    global info_line_count
    
    prompt += " > "
    user_input = input(f"{prompt}")
    info_line_count += prompt.count("\n") + 1

    delete_info_lines()
    return user_input

def title(title, level = 3):
    delete_info_lines()

    print(f"{'#' * level} {title}")
    print()

def default(message):
    delete_info_lines()
    print(message)
    print()

def response(character, response):
    delete_info_lines()
    dialogue = ""

    if isinstance(response, str):
        print_yellow(f"{character.name}:")
        print(response)
        dialogue = response
    else:
        print_cyan(f"{character.name}:")
        for chunk in response:
            if chunk.choices:
                if chunk.choices[0].delta.content:
                    dialogue += chunk.choices[0].delta.content
                    print(chunk.choices[0].delta.content, end='')
    print()

    return dialogue

def print_suggested_followups(followups):
    print()
    info("Suggested Follow-ups:", False)
    for i in range(settings.NUM_OF_SUGGESTED_FOLLOWUPS):
        if i < len(followups):
            info(f"\t{i + 1} - {followups[i]}", False)
    
def todo(message):
    delete_info_lines()
    
    print_cyan(f"TODO: {message}")
    call_stack = traceback.format_stack()
    print_grey(call_stack[-2].splitlines()[0])
    print()

def info(message, blank_line = True):
    global info_line_count

    print_grey(message)
    info_line_count += 1
    if blank_line:
        print()
        info_line_count += 1

def warning(message):
    delete_info_lines()
    
    print_yellow(f"WARNING: {message}")
    call_stack = traceback.format_stack()
    print_grey(call_stack[-2].splitlines()[0])
    print()

def error(message):
    delete_info_lines()
    
    print_red(f"ERROR: {message}")
    call_stack = traceback.format_stack()
    print_grey(call_stack[-2].splitlines()[0])
    print()

def delete_info_lines():
    global info_line_count

    for _ in range(info_line_count):
        sys.stdout.write('\x1b[1A')  # Move cursor up one line
        sys.stdout.write('\x1b[2K')  # Clear the line
        sys.stdout.flush()
    info_line_count = 0

def character_status(character):
    title(f"{character.name} Character Status", 2)
    todo("Output character status here")

def print_yellow(message):
    print(f"\033[93m{message}\033[0m")

def print_orange(message):
    print(f"\033[38;5;214m{message}\033[0m")

def print_red(message):
    print(f"\033[91m{message}\033[0m")

def print_green(message):
    print(f"\033[92m{message}\033[0m")

def print_blue(message):
    print(f"\033[94m{message}\033[0m")

def print_cyan(message):
    print(f"\033[96m{message}\033[0m")

def print_grey(message):
    print(f"\033[90m{message}\033[0m")

def open_for_editing(filepath):
    filepath = filepath.replace(" ", "\\ ").replace("(", "\\(").replace(")", "\\)").replace("'", "\\'")
    os.system(f"code {filepath}")

if __name__ == "__main__":
    print("Testing UI module...\n\n")
    
    title("Test Title 1", 1)
    title("Test Title 2", 2)
    title("Test Title 3", 3)
    title("Test Title 4", 4)
    title("Test Title 5", 5)
    
    todo("This is a test TODO message")

    info("This is an info message, it will be deleted after 2 seconds")
    time.sleep(2)
    delete_info_lines()

    warning("This is a warning message.")
    error("This is an error message.")

    print_suggested_followups(["Followup 1", "Followup 2", "Followup 3"])
    
    while True:
        user_input = get_user_input("Enter anything you like input (CTRL-C to quit)")    
        delete_info_lines()
        print(f"You entered: {user_input}")
