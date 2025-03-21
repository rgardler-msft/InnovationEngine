import os
import ui

from base_object import BaseObject

class Deployment(BaseObject):
    """
    Deployment class represents the deployment section of a document. It includes the necessary steps and configurations to deploy the documented workload in a specific environment.
    
    The deployment should be detailed and specific, providing all the necessary information to successfully deploy the workload.
    
    The deployment is the main part of the document. It is critical that the deployment is accurate and well-structured, as it sets the tone for the rest of the document.
    """

    def __init__(self):
        super().__init__()

    def get_prompt(self, document, auto = False):
        """
        Returns the user prompt for a documents deployment section.
        """
        content = ""
        for section in document.sections:
            content += section.content

        if not auto:
            prompt = ui.get_user_input(f"Provide any special instructions for the deployment section of the document titled {document.title}.\n")
            prompt += "\n\nCurrent content of the document:\n\n "
        else:
            prompt = "Write the deployment section for the document below:\n\n"

        prompt += content
        return prompt

if __name__ == "__main__":
    import ui
    from document import Document

    ui.title("Testing Deployment class...")
    
    ui.info("Create an empty Deployment")
    document = Document()
    document.title = "Test Document"
    
    deployment = Deployment()
    ui.info("Populate the Deployment")
    deployment.generate(document, True)
    deployment.display()

    ui.info("Saving Deployment to JSON file...")
    deployment.save("test.json")

    ui.info("Loading deployment from JSON file...")
    loaded_deployment = Deployment.load("test.json")

    assert deployment.meta_data == loaded_deployment.meta_data
    assert deployment.content == loaded_deployment.content

    ui.print_green("Loaded Deployment matches the original.")
    
    ui.info("Deleting test.json...")
    os.remove("test.json")
    deployment.delete()
    ui.info("Done.")