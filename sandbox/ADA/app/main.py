import json
import os
import settings
import ui
from document import Document

class DocumentIdeaManager:
    def __init__(self):
        self.validate_file(settings.DOCUMENT_CANDIDATES_FILEPATH)
        self.validate_file(settings.FAILED_DOCUMENT_GENERATIONS_FILEPATH)
        self.validate_file(settings.PASSED_DOCUMENT_GENERATIONS_FILEPATH)

    def validate_file(self, file_path):
        """
        Validates and updates the JSON file at the specified file path.
        This function reads a JSON file, validates each entry, and ensures that
        each entry contains the keys 'filename', 'title', and 'description'. If
        any of these keys are missing, they are added with default values:
        - 'filename': None
        - 'title': "TBC"
        - 'description': "TBC"
        The function also updates the UI with information about the validation
        process.
        Raises:
            FileNotFoundError: If the file at the specified path does not exist.
            json.JSONDecodeError: If the file contains invalid JSON.
        """

        directory = os.path.dirname(file_path)
        if os.path.exists(directory) and not os.path.exists(directory):
            os.makedirs(directory)
        if not os.path.exists(file_path):
            with open(file_path, 'w') as file:
                json.dump([], file)

        with open(file_path, 'r') as file:
            data = json.load(file)

        for entry in data:
            ui.delete_info_lines()
            ui.info(f"Validating entry: {entry['title']}")
            if 'filename' not in entry:
                entry['filename'] = None
                if 'tests_passed' not in entry:
                    entry['tests_passed'] = True
            if 'tests_passed' not in entry:
                entry['tests_passed'] = False
            if 'title' not in entry:
                entry['title'] = "TBC"
            if 'description' not in entry:
                entry['description'] = "TBC"
 
        with open(file_path, 'w') as file:
            json.dump(data, file, indent=4)


    def add_entry(self, title, description):
        entry = {"title": title, 
                 "description": description,
                 "filename": None
                }
        data = self._read_file(settings.DOCUMENT_CANDIDATES_FILEPATH)
        data.append(entry)
        self._write_file(data, settings.DOCUMENT_CANDIDATES_FILEPATH)

    def get_candidate_entries(self):
        """
        Retrieves candidate entries from a file.
        This method reads entries from a file and returns a list of entries
        where the 'filename' field is None. That is, it returns entries for
        which a document has not yet been generated.
        Returns:
            list: A list of candidate entries with 'filename' set to None.
        """

        return self._read_file(settings.DOCUMENT_CANDIDATES_FILEPATH)
        
    def get_generated_but_failed_entries(self):
        """
        Retrieves entries that have been generated but failed tests.
        This method reads entries from a file and returns a list of entries
        where the 'filename' field is not None and 'tests_passed' is False.
        Returns:
            list: A list of entries that have been generated but failed tests.
        """
        return self._read_file(settings.FAILED_DOCUMENT_GENERATIONS_FILEPATH)
        
    def get_generated_and_passed_entries(self):
        """
        Retrieves entries that have been generated and passed tests.
        This method reads entries from a file and returns a list of entries"
        where the 'filename' field is not None and 'tests_passed' is True.
        Returns:
            list: A list of entries that have been generated and passed tests.
        """
        return self._read_file(settings.PASSED_DOCUMENT_GENERATIONS_FILEPATH)
    
    def _read_file(self, file_path):
        with open(file_path, 'r') as file:
            return json.load(file)

    def _write_file(self, data, file_path):
        with open(file_path, 'w') as file:
            json.dump(data, file, indent=4)

def main():
    manager = DocumentIdeaManager()

    while True:
        choice = None
        while choice == None:
            ui.info("Menu:")

            idx = 0
            candidate_entries = manager.get_candidate_entries()
            if len(candidate_entries) > 0:
                ui.info("Generate a document from the ideas list:")
                for idx, entry in enumerate(candidate_entries, start=1):
                    ui.info(f"{idx} {entry['title']}\n\t{entry['description']}")

            failed_entries = manager.get_generated_but_failed_entries()
            if len(failed_entries) > 0:
                ui.info("Edit a document that failed tests:")
                for idx, entry in enumerate(failed_entries, start=idx + 1):
                    ui.info(f"{idx} {entry['title']}\n\t{entry['description']}")

            passed_entries = manager.get_generated_and_passed_entries()
            listGeneratedIdx = idx + 1
            ui.info(f"{listGeneratedIdx}. List {len(passed_entries)} generated documents")

            addItemIdx = listGeneratedIdx + 1
            ui.info(f"{addItemIdx}. Add a candidate document.")
            
            exitIdx = listGeneratedIdx + 1
            ui.info(f"{exitIdx}. Exit")

            try:
                choice = int(ui.get_user_input("Enter your choice: "))
            except ValueError:
                choice = None
                ui.warning("Invalid input. Please enter a number.")

        if choice == addItemIdx:
            title = ui.get_user_input("Enter title: ")
            description = ui.get_user_input("Enter description: ")
            manager.add_entry(title, description)
            ui.info("Entry added successfully.")
        elif choice == listGeneratedIdx:
            ui.info("Generated documents:")
            candidate_entries = manager.get_generated_and_passed_entries()
            if len(candidate_entries) == 0:
                ui.info("No generated documents available.")
            else:
                for idx, entry in enumerate(candidate_entries, start=1):
                    ui.info(f"{idx} {entry['title']}\n\t{entry['description']}")
            ui.get_user_input("Press Enter to continue...")
        elif choice == exitIdx:
            ui.info("Exiting...")
            break
        elif 1 <= choice <= len(candidate_entries):
            selected_entry = candidate_entries[choice - 1]
            ui.info(f"Generating document for: {selected_entry['title']}")
            document = Document(True, selected_entry['title'], selected_entry['description'])

            ui.info(f"Document saved as: {selected_entry['filename']}.md/.json")
            selected_entry['filename'] = document.filename() + ".md"
            
            if document.all_tests_passed():
                ui.print_green("Document generated and tested successfully.")
                selected_entry['tests_passed'] = True

                passed_entries = manager.get_generated_and_passed_entries()
                passed_entries.append(selected_entry)
                manager._write_file(passed_entries, settings.PASSED_DOCUMENT_GENERATIONS_FILEPATH)
                candidate_entries.pop(choice - 1)
                manager._write_file(candidate_entries, settings.DOCUMENT_CANDIDATES_FILEPATH)
            else:
                ui.warning("Document generated, but tests failed.")
                selected_entry['tests_passed'] = False
                
                failed_entries = manager.get_generated_but_failed_entries()
                failed_entries.append(selected_entry)
                manager._write_file(candidate_entries, settings.DOCUMENT_CANDIDATES_FILEPATH)
                candidate_entries.pop(choice - 1)
                manager._write_file(failed_entries, settings.FAILED_DOCUMENT_GENERATIONS_FILEPATH)        

                ui.info("Opening document for editing...")
                ui.open_for_editing(selected_entry['filename'])
        else:
            ui.warning("Invalid choice. Please try again.")

if __name__ == "__main__":
    main()