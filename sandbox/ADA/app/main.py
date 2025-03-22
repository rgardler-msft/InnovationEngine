import json
import os
import settings
import ui
from document import Document

class DocumentIdeaManager:
    def __init__(self, file_path):
        self.file_path = file_path
        if not os.path.exists(self.file_path):
            with open(self.file_path, 'w') as file:
                json.dump([], file)

    def add_entry(self, title, description):
        entry = {"title": title, "description": description}
        data = self._read_file()
        data.append(entry)
        self._write_file(data)

    def get_entries(self):
        return self._read_file()

    def _read_file(self):
        with open(self.file_path, 'r') as file:
            return json.load(file)

    def _write_file(self, data):
        with open(self.file_path, 'w') as file:
            json.dump(data, file, indent=4)

def main():
    manager = DocumentIdeaManager(settings.CANDIDATE_IDEAS_FILE)

    while True:
        ui.info("Menu:")

        idx = 0
        if (manager.get_entries() is not None):
            entries = manager.get_entries()
            ui.info("Generate a document from the ideas list:")
            for idx, entry in enumerate(entries, start=1):
                ui.info(f"{idx} {entry['title']}\n\t{entry['description']}")

        ui.info(f"{idx + 1}. Add an item")
        ui.info(f"{idx + 2}. Exit")

        choice = int(ui.get_user_input("Enter your choice: "))

        if choice == idx + 1:
            title = ui.get_user_input("Enter title: ")
            description = ui.get_user_input("Enter description: ")
            manager.add_entry(title, description)
            ui.info("Entry added successfully.")
        elif choice == idx + 2:
            ui.info("Exiting...")
            break
        elif choice <= idx:
            if 1 <= choice <= len(entries):
                selected_entry = entries[choice - 1]
                ui.info(f"Generating document for: {selected_entry['title']}")
                document = Document(True, selected_entry['title'], selected_entry['description'])
            else:
                ui.warning("Invalid document number.")
        else:
            ui.warning("Invalid choice. Please try again.")

if __name__ == "__main__":
    main()