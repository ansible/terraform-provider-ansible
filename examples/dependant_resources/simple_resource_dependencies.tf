# ===============================================
#  A simple exxample of dependant resources
# ===============================================

terraform {
  required_providers {
    ansible = {
      source  = "ansible/ansible"
      version = "~> 1.1.0"
    }
  }
}

# This resource will run example_resource_dependencies.yml with tag
# "set_vars", so that we'll only set variables 'var_a' and 'var_b'. 
# With this, the next resource "my_playbook_2" is made
# dependant on this resource; "my_playbook_1".
resource "ansible_playbook" "my_playbook_1" {
    playbook = "example_resource_dependencies.yml"

    # Temporary inventory config
    name  = "localhost"  # hostname - it's a required parameter

    # Play control
    tags = [
      "set_vars"
    ]

    extra_vars = {
      ansible_connection = "local"  # set to local for a localhost connection
    }

}


# This resource will run example_resource_dependencies.yml with tag
# "calc_sum", so that we'll only calculate the sum of variables
# "var_a" and "var_b" defined in the previous resource. This
# makes this resource dependant on the previous one.
resource "ansible_playbook" "my_playbook_2" {
    playbook = "./example_resource_dependencies.yml"

    # Temporary inventory config
    name  = "localhost"  # hostname - it's a required parameter

    # Play control
    tags = [
      "calc_sum"
    ]

    extra_vars = {
      ansible_connection = "local"  # set to local for a localhost connection
    }

    # Since this resource is dependant on resource "my_playbook_1"
    # we use "depends_on" list, which is a list of all resources
    # this resource depends on. In this case, it's only "my_playbook_1"
    # Removing this line will result in failure, since this resource
    # couldn't get the things that the previous resource created.
    depends_on = [ ansible_playbook.my_playbook_1 ]
}
