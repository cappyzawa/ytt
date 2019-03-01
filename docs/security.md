## Security

If you believe you have found a security issue in `ytt`, please responsibly disclose by contacting us at `dkalinin@pivotal.io`.

### Attack Vectors

- malicious template input
  - input tries to exhaust cpu/mem/disk resources
    - A: how does it affect go-yaml? ... https://en.wikipedia.org/wiki/Billion_laughs_attack
  - input tries to use YAML tagging to initialize custom objects (std yaml concern)
    - A: TBD

- malicious template code
  - code tries to load file contents from sensitive locations
    - A: templating is constrained to seeing only files explicitly specified by the user via -f flag (with possibly -R flag). unless user is tricked to provide sensitive file as input, template code is not able to access it. in other words, template runtime does not have facilities to access arbitrary filesystem locations.
  - code tries to exfiltrate data over network
    - A: template runtime does not have facilities to access network.
  - code tries to exhaust cpu/mem/disk resources
    - A: 
  - code tries to produce YAML that exhausts resources
    - A: TBD ... https://en.wikipedia.org/wiki/Billion_laughs_attack
  - meltdown/spectre style attacks
    - A: TBD

- CLI output directory
  - user is tricked to set --output directory to a sensitive filesystem location
    - A: template output is constrained to stdout or specified output directory via --output flag. if user is tricked to point --output directory to a sensitive filesystem location such as ~/.ssh/, attacker may be able to write templates (for example ~/.ssh/authorized_keys) that can be intepreted by the system as configuration/executable files.