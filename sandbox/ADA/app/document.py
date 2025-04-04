import importlib
import json
import jsonpickle
import os
import ui

from abc import ABC, abstractmethod
import requests

class Section:
    def __init__(self, name, requires_test = False):
        self.name = name
        self.requires_test = requires_test
        self.instance = None

class Document:
    def __init__(self, auto = False, title = None, description = None, published_source = None):
        self.auto = auto
        self.title = title
        self.description = description
        self.published_source = published_source
        self.meta_data = {}
        self.sections = [
            Section("Overview"), 
            Section("Deployment", True),
            Section("Summary")
        ]

        self.generate()

    def all_tests_passed(self):
        """
        Check if all tests have passed.
        """
        for section in self.sections:
            if section.requires_test and not section.instance.passed_tests:
                return False
        return True
    
    def get_errors(self):
        """
        Get the errors discovered in testing all sections. If no errors are found then an empty list is returned.
        If there all_tests_passed() is true then the last error (if any) in this list will have been resolved during generation.
        If all_tests_passed() is false then the last error in this list will be the most recent at which the generation was aborted.
        """
        errors = []
        for section in self.sections:
            if section.requires_test and section.instance.error_log:
                errors.extend(section.instance.error_log)
        return errors

    def filepath(self, override_name = None):
        if override_name is not None:
            name = override_name
        else:
            name = self.title
        encoded_name = "".join(c if c.isalnum() or c in (' ', '.', '_') else '_' for c in name).strip()
        return f"data/{self.__class__.__name__}/{encoded_name}"
    
    def filename(self, override_name = None):
        if override_name is not None:
            name = override_name
        else:
            name = self.title
        encoded_name = "".join(c if c.isalnum() or c in (' ', '.', '_') else '_' for c in name).strip()
        return f"{self.filepath(encoded_name)}/{encoded_name}"

    def save(self, filename = None):
        if filename is None:
            filename = self.filename()

        if not filename.endswith(".json"):
            filename = os.path.splitext(filename)[0] + ".json"
        
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
        if not filename.endswith(".json"):
            filename = os.path.splitext(filename)[0] + ".json"

        with open(filename, 'r') as f:
            state = json.load(f)

        object = jsonpickle.decode(state)

        return object
    
    def generate(self):
        if not self.title:
            self.title = ui.get_user_input("Enter the title of the document:")

        if self.published_source:
            response = requests.get(self.published_source)
            if response.status_code == 200:
                published_content = response.text
                if (self.description):
                    self.description += "\n\nCurrent content for this file is copied below. You should use this as a guide for the content you create, but should ensure that it is a valid executable doc and improve on readability wherever possible:\n\n" + published_content
                else:
                    self.description = "\n\nCurrent content for this file is copied below. You should use this as a guide for the content you create, but should ensure that it is a valid executable doc and improve on readability wherever possible:\n\n" + published_content
            else:
                raise Exception(f"Failed to retrieve content from {self.published_source}. Status code: {response.status_code}")
                return

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
    ui.info("Testing rewrite of existing content...")

    document = Document(True, "Create the infrastructure for deploying Apache Airflow on Azure Kubernetes Service (AKS)", None, "https://raw.githubusercontent.com/MicrosoftDocs/azure-aks-docs/refs/heads/main/articles/aks/airflow-create-infrastructure.md")
    document.generate()
    document.display()

    ui.info("Testing automatic creation of a document from a title and brief description...")

    document = Document(True, "Testing with RGs", "Create a resource group.")
    document.generate()
    document.display()

    ui.info("Testing manual creation of a Document class...")

    document = Document()
    document.generate()
    document.display()