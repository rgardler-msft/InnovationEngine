import os
import ui
from base_object import BaseObject

class Summary(BaseObject):
    """
    Summary class represents the summary of a document. It provides a high-level overview of the content and its main points.
    
    The summary should be concise and informative, giving the reader a clear understanding of the main points and structure of the document.
    """
    def __init__(self):
        super().__init__()
        self.title = "Summary"
        self.content = ""

    def get_prompt(self, document, auto = False):
        """
        Returns the user prompt for the summary of a document.
        """
        content = ""
        for section in document.sections:
            content += section.content

        prompt = "Write a summary section for the document below:\n\n"

        if not auto:
            prompt += ui.get_user_input(f"Any special instructions for the summary section?")
        
        prompt += "\n\n" + content
        return prompt
    
if __name__ == "__main__":
    import ui
    from document import Document

    ui.title("Testing Summary class...")
    
    ui.title("Create an empty Summary")
    document = Document()
    document.title = "Test Document"

    summary = Summary()
    ui.title("Populate the OverSummaryview")
    summary.generate(document, True)
    summary.display()

    ui.title("Saving Summary to JSON file...")
    summary.save("test.json")

    ui.title("Loading summary from JSON file...")
    loaded_summary = Summary.load("test.json")

    assert summary.meta_data == loaded_summary.meta_data
    assert summary.content == loaded_summary.content

    ui.print_green("Loaded Summary matches the original.")
    
    ui.info("Deleting test.json...")
    os.remove("test.json")
    ui.info("Done.")