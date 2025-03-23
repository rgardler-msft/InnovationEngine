import os
import ui

from base_object import BaseObject

class Overview(BaseObject):
    """
    Overview class represents the overview of a document. It summarizes the content and provides a high-level view of the document.

    The overview should be concise and informative, giving the reader a clear understanding of the main points and structure of the document.

    The overview is the first part of the document to be created and is used to inform the generation of detail sections. It is critical that the overview is accurate and well-structured, as it sets the tone for the rest of the document.
    """

    def __init__(self):
        super().__init__()

    def get_prompt(self, document):
        """
        Returns the user prompt for the overview of a document.
        """
        content = f"# {document.title}\n\n"
        for section in document.sections:
                if section.instance and section.instance.content:
                    content += section.instance.content

        if not document.auto and not document.description:
            prompt = ui.get_user_input(f"Provide any guidance you have for the agent creating your document '{document.title}'.\n")
        else:
            if document.description:
                prompt = f"Write an overview for a document with the title '{document.title}' and a description of content of '{document.description}'.\n\n"
            else:
                prompt = f"Write an overview for a document with the title '{document.title}'.\n\n"
            
        if content:
            prompt += "Current content of the document is:\n\n " + content
        return prompt

if __name__ == "__main__":
    import ui
    from document import Document

    ui.title("Testing Overview class...")
    
    ui.title("Create an empty Overview")
    document = Document()
    document.title = "Test Document"
    
    overview = Overview()
    ui.title("Populate the Overview")
    overview.generate(Document, True)
    overview.display()

    ui.title("Saving Overview to JSON file...")
    overview.save("test.json")

    ui.title("Loading overview from JSON file...")
    loaded_overview = Overview.load("test.json")

    assert overview.meta_data == loaded_overview.meta_data
    assert overview.content == loaded_overview.content

    ui.print_green("Loaded Overview matches the original.")
    
    ui.info("Deleting test.json...")
    os.remove("test.json")
    ui.info("Done.")