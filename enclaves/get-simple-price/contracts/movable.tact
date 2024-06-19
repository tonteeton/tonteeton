import "@stdlib/ownable";
import "./oracleProtocol.tact";

message (0x1af44a9f) MoveTo {
    newAddress: Address;
    moveCompleted: Bool;
}

trait  Movable with Ownable {

    owner: Address;
    prevAddress: Address?;
    newAddress: Address;
    moved: Bool;

    receive (msg: MoveTo) {
        self.requireOwner();
        self.requireNotMoved();

        self.newAddress = msg.newAddress;
        let sb: StringBuilder = beginString();
        sb.append("Start moving to: ");
        sb.append(self.newAddress.toString());
        emit(sb.toString().asComment());

        if (msg.moveCompleted) {
            send(SendParameters{
                to: self.newAddress,
                value: 0,
                mode: SendRemainingValue,
                bounce: false,
                body: "MoveConfirmation".asComment(),
            });
        }
    }

    receive ("MoveConfirmation") {
        require(sender() == self.prevAddress, "Previous address required");
        self.reply("MoveCompleted".asComment());
    }

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

    receive ("NewAddress") {
        self.replyNewAddress(0);
    }

    get fun newAddress() : Address {
        return self.newAddress;
    }

    fun replyNewAddress(queryId: Int) {
        self.reply(OracleNewAddressResponse{
            newAddress: self.newAddress,
            queryId: queryId,
        }.toCell());
    }

    fun requireNotMoved() {
        require(!self.moved, "Moved to new address");
    }

}
