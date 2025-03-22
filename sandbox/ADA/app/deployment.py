import os
import subprocess
import ui

from base_object import BaseObject

class Deployment(BaseObject):
    """
    Deployment class represents the deployment section of a document. It includes the necessary steps and configurations to deploy the documented workload in a specific environment.
    
    The deployment should be detailed and specific, providing all the necessary information to successfully deploy the workload.
    
    The deployment is the main part of the document. It is critical that the deployment is accurate and well-structured, as it sets the tone for the rest of the document.
    """
    passed_tests = False

    def __init__(self):
        super().__init__()

    def generate(self, document, auto = False):
        """
        Generates the deployment section of the document.
        
        Args:
            document (Document): The document to which this deployment section belongs.
            auto (bool): If True, the prompt is generated automatically without user input. Note that some sections always automatically generate a prompt.
        """
        super().generate(document, auto)
        self.auto_fix_errors()

    def get_prompt(self, document):
        """
        Returns the user prompt for a documents deployment section.
        """
        content = ""
        for section in document.sections:
            content += section.content

        if not document.auto:
            prompt = ui.get_user_input(f"Provide any special instructions for the deployment section of the document titled {document.title}.\n")
            prompt += "\n\nCurrent content of the document:\n\n "
        else:
            prompt = "Write the deployment section for the document below:\n\n"

        prompt += content
        return prompt
    
    def auto_fix_errors(self, max_tests=10):
        """
        Test the Document in the Innovation Engine.
        If the test fails then another call to the LLM is made to try to fix the error.
        The resulting document is tested again.
        This cycle will be repeated until the test passes or the maximum number of attempts is reached.

        Args:
            max_tests (int): The maximum number of tests to run before giving up.
        returns:
            bool: True if the test passed, False otherwise. Note this can also be retrieved from `Deployment.passed_tests`.
        """
        i = 0
        while i < max_tests and self.passed_tests == False:
            i += 1
            ui.info(f"Testing the generated content in IE (attempt {i} of {max_tests})")
            filename = self.filename()

            try:
                process = subprocess.Popen(f"ie test '{filename}.md'", shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                for line in iter(process.stdout.readline, ''):
                    print(line, end='')  # Stream to stdout
                process.stdout.close()
                process.wait()
                if process.returncode != 0:
                    raise subprocess.CalledProcessError(process.returncode, process.args, output=process.stderr.read())
                self.passed_tests = True
            except subprocess.CalledProcessError as e:
                # TODO: We need a better output on syserror - see https://github.com/Azure/InnovationEngine/issues/258
                error_message = e.output
                ui.error(f"Error executing the document: {error_message}")
                self.user_edits(self.__class__.__name__, f"Fix this error, thrown when executing the document.\n\n{error_message}")
        
        return self.passed_tests

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
    
    deployment.auto_fix_errors()

    print()
    deployment.delete()
    
    ui.info("Done.")