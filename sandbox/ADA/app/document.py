import importlib
import json
import jsonpickle
import os
import ui

from abc import ABC, abstractmethod

class Section:
    def __init__(self, name, requires_test = False):
        self.name = name
        self.requires_test = requires_test
        self.instance = None

class Document:
    def __init__(self, auto = False, title = None, description = None):
        self.auto = auto
        self.title = title
        self.description = description
        self.meta_data = {}
        self.sections = [Section("Overview"), 
                                  Section("Deployment", True),
                                  Section("Summary")]

        self.generate()

    def all_tests_passed(self):
        """
        Check if all tests have passed.
        """
        for section in self.sections:
            if section.requires_test and not section.instance.passed_tests:
                return False
        return True
    
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

    def save(self, filename = None):
        if filename is None:
            filename = self.filename()
        
        os.makedirs(os.path.dirname(filename), exist_ok=True)

        state = jsonpickle.encode(self)
        with open(filename + ".json", 'w') as f:
            json.dump(state, f)

        # Strip the current extension and add .md
        base_filename, _ = os.path.splitext(filename)
        filename = f"{base_filename}.md"

        content = f"# {self.title}\n\n"
        for section in self.sections:
            if section.instance and section.instance.content:
                content += section.instance.content + "\n\n"

        with open(filename, 'w') as f:
            f.write(content)

    @classmethod
    def load(cls, filename):
        with open(filename, 'r') as f:
            state = json.load(f)

        object = jsonpickle.decode(state)

        return object
    
    def generate(self):
        if not self.title:
            self.title = ui.get_user_input("Enter the title of the document:")

        for section in self.sections:
            module_name = section.name.lower()
            module = importlib.import_module(module_name)
        
            cls = getattr(module, section.name)
            instance = cls()
            instance.generate(self)

            section.instance = instance

        self.save(self.filename())
        ui.open_for_editing(f"{self.filename()}.md")

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
            ext = f"_{type}.json"
        else:
            ext = f"_{type}.md"
        return ext

    def __str__(self):
        content = f"# {self.title}\n\n"

        for section in self.sections:
            content += f"{section}\n\n"

    def display(self):
        """
        Display the current state of the Document.
        """
        ui.title(self.title, 1)
        for section in self.sections:
            section.instance.display()

if __name__ == "__main__":
    ui.info("Testing automatic creation of a Document class...")

    document = Document(True, "Testing with RGs", "Create a resource group.")
    document.generate()
    document.display()

    ui.info("Testing manual creation of a Document class...")

    document = Document()
    document.generate()
    document.display()