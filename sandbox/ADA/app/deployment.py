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
    document.title = "Create a basic Linux VM on Azure"
    
    deployment = Deployment()
    ui.info("Create the Deployment section")
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
    
    max_tests = 10
    passed_tests = False
    i = 0
    while i < max_tests and passed_tests == False:
        i += 1
        ui.info(f"Testing the generated content in IE (attempt {i} of {max_tests})")
        filename = deployment.filename()
        import subprocess

        try:
            result = subprocess.run(f"ie test '{filename}.md'", shell=True, check=True, capture_output=True, text=True)
            output = result.stdout
            print(output)
            passed_tests = True
        except subprocess.CalledProcessError as e:
            # TODO: We need a better output on syserror - see https://github.com/Azure/InnovationEngine/issues/258
            error_message = e.output
            ui.error(f"Error executing the document: {error_message}")
            deployment.user_edits(deployment.__class__.__name__, f"Attempt to fix this error when executing the document.\n\n{error_message}")
    
    print()
    deployment.delete()
    
    ui.info("Done.")