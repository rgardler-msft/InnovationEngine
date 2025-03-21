import importlib
import json
import jsonpickle
import os
import llm
import settings
import ui

from abc import ABC, abstractmethod

class BaseObject:
    def __init__(self, title = None):
        self.title = title
        self.meta_data = {}
        self.content = ""

        if (title is not None):
            self.configure_llm_prompts()

    def filepath(self, override_name = None):
        if override_name is not None:
            name = override_name
        else:
            name = self.title
        return f"data/{self.__class__.__name__}/{name}"
    
    def filename(self, override_name = None):
        if override_name is not None:
            name = override_name
        else:
            name = self.title
        return f"{self.filepath(name)}/{name}"

    @property
    def prompt_folder(self):
        return f"{settings.ROOT_PROMPT_FOLDER}"

    def save(self, filename = None):
        if filename is None:
            filename = self.filename()
        
        state = jsonpickle.encode(self)
        with open(filename, 'w') as f:
            json.dump(state, f)

    @classmethod
    def load(cls, filename):
        with open(filename, 'r') as f:
            state = json.load(f)

        object = jsonpickle.decode(state)
        
        return object
    
    def delete(self):
        if os.path.exists(self.filename()):
            os.remove(self.filename())
    
    def generate(self, document, auto = False):
        type = self.__class__.__name__
        if self.title is None:
            self.title = f"{type} - {document.title}"

        prompt = self.get_prompt(document, auto)

        data = None

        if (self.title is not None):
            data = self.load_section_data(type)
    
        if data is None or len(data) == 0:
            messages = []
            messages.append(llm.get_system_message_from_file(f"{self.prompt_folder}/{type}_system_prompt.txt"))
            messages.append({
                "role": "user", 
                "content": f"Generate a {type}. \n{prompt}."
            })
            response = llm.send_message(messages, False)

            if isinstance(response, str):
                data = None
            else:
                data = ""
                for chunk in response:
                    if chunk.choices:
                        if chunk.choices[0].delta.content:
                            data += chunk.choices[0].delta.content

        if data is not None:
            self.content = data
            self.save_section_data(type, self.content)
        
        if not auto:
            while (self.user_edits(type)):
                pass

        self.content = self.load_section_data(type)

    @abstractmethod
    def get_prompt(self, document = None, auto = False):
        """
        Returns the user prompt to pass to the LLM to generate the content.

        This function should be implemented by subclasses to provide the specific prompt for the type of object.
        
        Args:
            document (Document): A Document object that contains the content of the document so far.
            auto (bool): If True, the prompt is generated automatically without user input. Note that some sections always automatically generate a prompt.
        Returns:
            str: The prompt to be used for generating content.
        """
        pass

    def user_edits(self, section_name):
        """
        Enables the editing of a prompt by the user and processes the changes.
        This function allows the user to edit a metadata file or provide a prompt for desired changes.
        It then processes the user's input, updates the metadata, and handles any necessary file operations.

        Note that it is possible that the user has edited the save file for this section. This function will not check for that
        Returns:
            bool: True if the metadata was edited either by an LLM prompt or in a way that causes the file to move, False otherwise.
        """
        
        filename = f"{self.filename()}{self.get_extension(section_name)}"
        ui.open_for_editing(filename)
        ui.delete_info_lines()
        prompt = ui.get_user_input(f"{section_name} created and opened in the editor. You can now take one of three actions:\n\t1) Edit the file, save, hit enter here to proceed.\n\t2) Type a prompt for desired changes here and hit enter.\n\t3) Continue without changes (press enter).\n\n")

        ui.delete_info_lines()
    
        original_name = self.title
        if section_name == "meta_data":
            content = json.dumps(self.meta_data)
        else:
            content = self.load_section_data(section_name)
        is_edited = False
        if prompt is None or len(prompt) > 0:
            is_edited = True
            ui.default(prompt)

            system_prompt = llm.get_system_message_from_file(f"{self.prompt_folder}/{section_name}_system_prompt.txt")
            
            prompt = f"{system_prompt}\n\nEdit the content below to reflect the following changes:\n\n{prompt}\n\n{content}"
            messages = []
            messages.append(llm.get_system_message_from_file(f"{self.prompt_folder}/{section_name}_system_prompt.txt"))
            messages.append({
                    "role": "user", 
                    "content": prompt
                })
            response = llm.send_message(messages, False)
            
            data = ""
            for chunk in response:
                if chunk.choices:
                    if chunk.choices[0].delta.content:
                        data += chunk.choices[0].delta.content
            
            if (section_name == "meta_data"):
                self.meta_data = json.loads(data)
                self.title = self.meta_data["name"]
                self.save_section_data(section_name, json.dumps(self.meta_data))
            else:
                self.save_section_data(section_name, data)
            ui.open_for_editing(f"{self.filename()}{self.get_extension(section_name)}")

        if (section_name == "meta_data"):
            self.meta_data = json.loads(self.load_section_data("meta_data"))
            if (original_name != self.meta_data["name"]):
                is_edited = True

                ui.info(f"Name changed from {original_name} to {self.meta_data['name']}. Moving Files...")

                os.remove(f"{self.filename(original_name)}{self.get_extension(section_name)}")
                parent_dir = os.path.dirname(self.filename(original_name))
                if not os.listdir(parent_dir):
                    os.rmdir(parent_dir)
                
                self.title = self.meta_data["name"]
                self.save_section_data(section_name, json.dumps(self.meta_data))
                ui.open_for_editing(f"{self.filename()}{self.get_extension(section_name)}")

        return is_edited

    def load_section_data(self, type):
        ext = self.get_extension(type)

        if (os.path.exists(self.filename() + ext)):
            with open(self.filename() + ext, "r") as file:
                return file.read()
        return None

    def save_section_data(self, type, data):
        ext = self.get_extension(type)
        if (ext.endswith(".json")):
            data = json.dumps(json.loads(data), indent=2)

        os.makedirs(os.path.dirname(self.filename()), exist_ok=True)
        with open(self.filename() + ext, "w") as file:
            file.write(data)

    def get_extension(self, type):
        if type == "meta_data":
            ext = f".json"
        else:
            ext = f".md"
        return ext

    def __str__(self):
        return f"{self.title}"

    def display(self):
        """
        Display the current state of the Object.
        """
        ui.default(f"{self.content}")

