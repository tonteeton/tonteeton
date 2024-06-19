// Movable is a trait that provides functionality for contract updates.
// It allows the contract to be moved to a new address.
import "@stdlib/ownable";
import "./oracleProtocol.tact";

// MoveTo is a message structure used to initiate the contract move to a new address.
// The "moveCompleted" flag indicates whether the move operation has been permanently completed,
// making the old contract unusable and preventing further moves.
message (0x1af44a9f) MoveTo {
    newAddress: Address;
    moveCompleted: Bool;
}

trait  Movable with Ownable {

    owner: Address;
    prevAddress: Address?; // Previous address of the contract, if it is not the first version.
    newAddress: Address;   // Address where the contract has been moved, or the current contract address.
    moved: Bool;           // Indicates if the contract has already been moved.

    get fun newAddress() : Address {
        return self.newAddress;
    }

    // Handles the MoveTo message to initiate the contract move.
    receive (msg: MoveTo) {
        self.requireOwner();
        self.requireNotMoved();

        self.newAddress = msg.newAddress;
        let sb: StringBuilder = beginString();
        sb.append("Start moving to: ");
        sb.append(self.newAddress.toString());
        emit(sb.toString().asComment());

        if (msg.moveCompleted) {
            // Sends a confirmation request to the new contract.
            send(SendParameters{
                to: self.newAddress,
                value: 0,
                mode: SendRemainingValue,
                bounce: false,
                body: "MoveConfirmation".asComment(),
            });
        }
    }

    // Confirms to the previous contract that the new contract is ready.
    receive ("MoveConfirmation") {
        require(sender() == self.prevAddress, "Previous address required");
        self.reply("MoveCompleted".asComment());
    }

    // Finalize the contract move.
    receive ("MoveCompleted") {
        require(!self.moved, "Already moved");
        require(self.newAddress != myAddress(), "Address is not new");
        require(sender() == self.newAddress, "New address sender expected");

        self.moved = true;

        let sb: StringBuilder = beginString();
        sb.append("Moved to: ");
        sb.append(self.newAddress.toString());
        emit(sb.toString().asComment());
    }

    // Respond with the new address of the contract.
    receive ("NewAddress") {
        self.replyNewAddress(0);
    }

    fun replyNewAddress(queryId: Int) {
        self.reply(OracleNewAddressResponse{
            newAddress: self.newAddress,
            queryId: queryId,
        }.toCell());
    }

    // Checks that the contract hasn't moved yet.
    fun requireNotMoved() {
        require(!self.moved, "Moved to new address");
    }

}
